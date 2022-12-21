package handler

import (
	"context"

	"github.com/Yosorable/ms-metadata/core/model"
	"github.com/Yosorable/ms-metadata/global"
	"github.com/Yosorable/ms-shared/protoc_gen/common"
	pb "github.com/Yosorable/ms-shared/protoc_gen/metadata"
	"github.com/Yosorable/ms-shared/utils"
)

func GetObj(ctx context.Context, req *pb.GetObjRequest) (*pb.GetObjReply, error) {

	single := true // req中查询参数是id或者name，不是list
	if s, err := checkGetObjReq(req); err != nil {
		return nil, err
	} else {
		single = s
	}

	var objs []*model.MtObj
	fieldsMap := make(map[int][]*pb.ObjFieldDao) // obj_id: fields
	if req.GetQueryType() != pb.GetObjRequest_FIELDS {
		switch {
		case req.Id != nil:
			var err error
			objs, err = getObjByIDs([]int32{*req.Id})
			if err != nil {
				return nil, utils.NewStatusError(7000, err)
			}
		case len(req.IdList) != 0:
			var err error
			objs, err = getObjByIDs(req.IdList)
			if err != nil {
				return nil, utils.NewStatusError(7000, err)
			}
		case req.Name != nil:
			var err error
			objs, err = getObjByNames([]string{*req.Name})
			if err != nil {
				return nil, utils.NewStatusError(7000, err)
			}
		case len(req.NameList) != 0:
			var err error
			objs, err = getObjByNames(req.NameList)
			if err != nil {
				return nil, utils.NewStatusError(7000, err)
			}
		}

		res := &pb.GetObjReply{}
		if req.GetQueryType() == pb.GetObjRequest_OBJ {
			if single {
				if len(objs) == 1 {
					obj := objs[0]
					timeInfo, des := getObjTimeInfoAndDescription(req, obj)
					res.ObjWithFields = &pb.ObjDaoWithFieldsDao{
						Obj: &pb.ObjDao{
							Id:          int32(obj.ID),
							Name:        obj.Name,
							Description: des,
							TimeInfo:    timeInfo,
						},
					}
				}
				return res, nil
			}
			for _, obj := range objs {
				timeInfo, des := getObjTimeInfoAndDescription(req, obj)
				res.ObjWithFieldsList = append(res.ObjWithFieldsList, &pb.ObjDaoWithFieldsDao{
					Obj: &pb.ObjDao{
						Id:          int32(obj.ID),
						Name:        obj.Name,
						Description: des,
						TimeInfo:    timeInfo,
					},
				})
			}
			return res, nil
		}
	} else {
		var objIds []int
		switch {
		case req.Id != nil:
			objIds = append(objIds, int(*req.Id))
		case len(req.IdList) != 0:
			for _, id := range req.IdList {
				objIds = append(objIds, int(id))
			}
		case req.Name != nil:
			res, err := getObjIDsByNames([]string{*req.Name})
			if err != nil {
				return nil, err
			}
			objIds = res
		case len(req.NameList) != 0:
			res, err := getObjIDsByNames(req.NameList)
			if err != nil {
				return nil, err
			}
			objIds = res
		}
		for _, id := range objIds {
			res, err := getObjFieldsDto(req, id)
			if err != nil {
				return nil, err
			}
			fieldsMap[id] = res
		}

		res := &pb.GetObjReply{}
		if single {
			if len(objIds) == 1 {
				res.ObjWithFields = &pb.ObjDaoWithFieldsDao{
					Obj:    &pb.ObjDao{Id: int32(objIds[0])},
					Fields: fieldsMap[objIds[0]],
				}
			}
			return res, nil
		}
		for k, v := range fieldsMap {
			res.ObjWithFieldsList = append(res.ObjWithFieldsList, &pb.ObjDaoWithFieldsDao{
				Obj:    &pb.ObjDao{Id: int32(k)},
				Fields: v,
			})
		}
		return res, nil
	}

	for _, obj := range objs {
		fields, err := getObjFieldsDto(req, obj.ID)
		if err != nil {
			return nil, err
		}
		fieldsMap[obj.ID] = fields
	}

	res := &pb.GetObjReply{}

	if single {
		if len(objs) == 1 {
			obj := objs[0]
			timeInfo, des := getObjTimeInfoAndDescription(req, obj)
			res.ObjWithFields = &pb.ObjDaoWithFieldsDao{
				Obj: &pb.ObjDao{
					Id:          int32(obj.ID),
					Name:        obj.Name,
					Description: des,
					TimeInfo:    timeInfo,
				},
				Fields: fieldsMap[obj.ID],
			}
		}
		return res, nil
	}

	for _, obj := range objs {
		timeInfo, des := getObjTimeInfoAndDescription(req, obj)
		res.ObjWithFieldsList = append(res.ObjWithFieldsList, &pb.ObjDaoWithFieldsDao{
			Obj: &pb.ObjDao{
				Id:          int32(obj.ID),
				Name:        obj.Name,
				Description: des,
				TimeInfo:    timeInfo,
			},
			Fields: fieldsMap[obj.ID],
		})
	}

	return res, nil
}

func checkGetObjReq(req *pb.GetObjRequest) (single bool, err error) {
	if req == nil {
		return single, utils.NewStatusError(7000, "please set request")
	}

	setCnt := 0
	if req.Id != nil {
		setCnt++
		single = true
	}
	if req.Name != nil {
		setCnt++
		single = true
	}
	if len(req.IdList) != 0 {
		setCnt++
	}
	if len(req.NameList) != 0 {
		setCnt++
	}
	if setCnt != 1 {
		return single, utils.NewStatusError(7000, "you set too many fields")
	}

	return
}

func getObjByIDs(ids []int32) (objs []*model.MtObj, err error) {
	db := global.DATABASE
	err = db.Where("id in ?", ids).Find(&objs).Error
	return
}

func getObjByNames(names []string) (objs []*model.MtObj, err error) {
	db := global.DATABASE
	err = db.Where("name in ?", names).Find(&objs).Error
	return
}

func getObjIDsByNames(names []string) (ids []int, err error) {
	db := global.DATABASE
	err = db.Model(&model.MtObj{}).Select("id").Where("name in ?", names).Find(&ids).Error
	return
}

func getObjTimeInfoAndDescription(req *pb.GetObjRequest, obj *model.MtObj) (*common.TimeInfo, *string) {
	if req.GetShowTimeinfoAndDescription() {
		return &common.TimeInfo{
			CreatedAt: obj.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: obj.UpdatedAt.Format("2006-01-02 15:04:05"),
		}, obj.Description
	}
	return nil, nil
}

func getObjFieldsDto(req *pb.GetObjRequest, objId int) (res []*pb.ObjFieldDao, err error) {
	var fields []*model.MtObjField
	db := global.DATABASE
	err = db.Where("obj_id = ?", objId).Find(&fields).Error

	if err != nil {
		return nil, err
	}

	for _, ele := range fields {
		timeInfo, des := getFieldTimeInfoAndDescription(req, ele)
		res = append(res, &pb.ObjFieldDao{
			Id:        int32(ele.ID),
			ObjId:     int32(ele.ObjID),
			Name:      ele.Name,
			FieldType: pb.ObjField_FieldType(ele.FieldType),
			LookupObjId: func() *int32 {
				if ele.LookupObjID != nil {
					res := int32(*ele.LookupObjID)
					return &res
				}
				return nil
			}(),
			Description: des,
			IsAllowNull: ele.IsAllowNull,
			TimeInfo:    timeInfo,
		})
	}
	return
}

func getFieldTimeInfoAndDescription(req *pb.GetObjRequest, field *model.MtObjField) (*common.TimeInfo, *string) {
	if req.GetShowTimeinfoAndDescription() {
		return &common.TimeInfo{
			CreatedAt: field.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: field.UpdatedAt.Format("2006-01-02 15:04:05"),
		}, field.Description
	}
	return nil, nil
}
