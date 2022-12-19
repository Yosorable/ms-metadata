package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Yosorable/ms-metadata/core/model"
	mtutils "github.com/Yosorable/ms-metadata/core/utils"
	"github.com/Yosorable/ms-metadata/global"
	pb "github.com/Yosorable/ms-shared/protoc_gen/metadata"
	"github.com/Yosorable/ms-shared/utils"
	"gorm.io/gorm"
)

func CreateObj(ctx context.Context, req *pb.CreateObjRequest) (*pb.CreateObjReply, error) {
	if err := checkCreateRequest(req); err != nil {
		return nil, err
	}

	mtObj := &model.MtObj{
		Name:        req.GetName(),
		Description: req.Description,
	}

	mtFields := []*model.MtObjField{}
	for _, ele := range req.GetFields() {
		mtFields = append(mtFields, &model.MtObjField{
			Name:        ele.GetName(),
			FieldType:   int(ele.GetFieldType()),
			Description: ele.Description,
			LookupObjID: func() *int {
				if ele.LookupObjId != nil {
					res := int(*ele.LookupObjId)
					return &res
				}
				return nil
			}(),
			IsAllowNull: ele.GetIsAllowNull(),
		})
	}

	db := global.DATABASE

	var tableCnt int64
	cntErr := db.Raw(`
    SELECT count(*)
    FROM information_schema.tables
    WHERE 
        table_schema = '` + global.CONFIG.MySQL.DBName + `'
        AND
         table_name = '` + req.GetName() + `'
    LIMIT 1;
    `).Count(&tableCnt).Error
	if cntErr != nil {
		return nil, utils.NewStatusError(7000, cntErr)
	}
	if tableCnt > 0 {
		return nil, utils.NewStatusError(7000, "table already exists")
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// 必须先创建表，再存储记录，否则建表错误回滚不会删除记录
		// create obj table
		pfn := mtutils.PackFieldNameForSql
		sql := `CREATE TABLE ` + pfn(req.GetName()) + ` (`
		sql += "\n"

		idName := ""
		for _, ele := range req.GetFields() {
			sql += generateFieldSQL(ele) + ",\n"
			if ele.FieldType == pb.ObjField_ID {
				idName = ele.Name
			}
		}
		sql += `
            ` + mtutils.PackFieldNameForSql("created_at") + ` 
                datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0),
            ` + mtutils.PackFieldNameForSql("updated_at") + ` 
                datetime(0) NOT NULL DEFAULT CURRENT_TIMESTAMP(0) 
                    ON UPDATE CURRENT_TIMESTAMP(0),`
		sql += "\n"
		if idName != "" {
			sql += fmt.Sprintf("PRIMARY KEY (`%s`),\n", idName)
		}

		for _, ele := range req.GetObjIndex() {
			sql += generateFieldIndexSQL(ele) + ",\n"
		}

		sql = strings.TrimSuffix(sql, ",\n")
		sql += "\n);"
		if txErr := tx.Exec(sql).Error; txErr != nil {
			return txErr
		}

		// create metadata
		if txErr := tx.Create(&mtObj).Error; txErr != nil {
			return txErr
		}

		for _, ele := range mtFields {
			ele.ObjID = mtObj.ID
		}

		if txErr := tx.Create(&mtFields).Error; txErr != nil {
			return txErr
		}

		fieldIdsMap := make(map[string]int, len(mtFields))
		for _, ele := range mtFields {
			fieldIdsMap[ele.Name] = ele.ID
		}
		if len(req.GetObjIndex()) > 0 {
			mtObjIndexs := []*model.MtObjIndex{}
			for _, ele := range req.GetObjIndex() {
				fieldIds := []int{}
				for _, name := range ele.GetObjFieldNames() {
					fieldIds = append(fieldIds, fieldIdsMap[name])
				}
				bytes, jsonErr := json.Marshal(fieldIds)
				if jsonErr != nil {
					return jsonErr
				}
				mtObjIndexs = append(mtObjIndexs, &model.MtObjIndex{
					ObjFieldIds: string(bytes),
					IndexType:   int(ele.GetIndexType()),
				})
			}
			if txErr := tx.Create(&mtObjIndexs).Error; txErr != nil {
				return txErr
			}
		}

		return nil
	}); err != nil {
		_ = db.Exec("DROP table " + mtutils.PackFieldNameForSql(req.GetName())).Error
		return nil, utils.NewStatusError(7000, err)
	}

	return &pb.CreateObjReply{}, nil
}

func checkCreateRequest(req *pb.CreateObjRequest) error {
	if !mtutils.CheckFieldName(req.GetName()) {
		return utils.NewStatusError(7001, "obj name not allow")
	}
	if len(req.GetFields()) == 0 {
		return utils.NewStatusError(7002, "you should define obj fields")
	}
	fieldsIndex := map[string]int{}
	idCnt := 0
	for i, field := range req.GetFields() {
		switch {
		case !mtutils.CheckFieldName(field.Name):
			return utils.NewStatusError(7003, "field name not allow")
		case field.FieldType == pb.ObjField_UNSET_NOTALLOW:
			return utils.NewStatusError(7004, "field type should be set correctly")
		case func() bool {
			_, exists := fieldsIndex[field.Name]
			return exists
		}():
			return utils.NewStatusError(7005, "field name repeated")
		}
		if field.FieldType == pb.ObjField_ID {
			idCnt++
			if idCnt > 1 {
				return utils.NewStatusError(7006, "you can only set one id")
			}
		}
		fieldsIndex[field.Name] = i
	}

	for _, ele := range req.GetObjIndex() {
		switch {
		case len(ele.ObjFieldNames) == 0:
			return utils.NewStatusError(7007, "index fields should be set")
		case ele.IndexType == pb.ObjIndex_UNSET_NOTALLOW:
			return utils.NewStatusError(7008, "index type should be set correctly")
		}

		for _, fieldName := range ele.GetObjFieldNames() {
			if _, exists := fieldsIndex[fieldName]; !exists {
				return utils.NewStatusError(7009, "index field name should be set correctly")
			}
		}
	}
	return nil
}

func generateFieldSQL(field *pb.ObjField) (res string) {
	res += mtutils.PackFieldNameForSql(field.Name) + " "
	switch field.FieldType {
	case pb.ObjField_ID:
		res += `int UNSIGNED NOT NULL AUTO_INCREMENT`
		return
	case pb.ObjField_LOOKUPID:
		res += `int UNSIGNED NOT NULL`
		return
	case pb.ObjField_INTEGER:
		res += `int`
	case pb.ObjField_DOUBLE:
		res += `double`
	case pb.ObjField_STRING:
		res += `varchar(255)`
	case pb.ObjField_TEXT:
		res += `text`
	case pb.ObjField_DATETIME:
		res += `datetime`
	}
	res += ` `
	if field.GetIsAllowNull() {
		res += `NULL `
	} else {
		res += `NOT NULL `
	}
	return
}

func generateFieldIndexSQL(index *pb.ObjIndex) (res string) {
	switch index.GetIndexType() {
	case pb.ObjIndex_COMMON:
		res += "INDEX "
	case pb.ObjIndex_UNIQUE:
		res += "UNIQUE INDEX "
	}
	indexName := ""
	indexFields := "("
	for _, ele := range index.GetObjFieldNames() {
		indexName += ele + "_"
		indexFields += mtutils.PackFieldNameForSql(ele) + ","
	}
	indexFields = strings.TrimSuffix(indexFields, ",") + ")"
	indexName += "index"
	res += mtutils.PackFieldNameForSql(indexName) + indexFields
	return
}
