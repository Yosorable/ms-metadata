package handler

import (
	"context"

	"github.com/Yosorable/ms-metadata/core/model"
	"github.com/Yosorable/ms-metadata/global"
	"github.com/Yosorable/ms-shared/protoc_gen/common"
	pb "github.com/Yosorable/ms-shared/protoc_gen/metadata"
	"github.com/Yosorable/ms-shared/utils"
)

func GetObjByID(ctx context.Context, req *pb.GetObjRequest) (*pb.GetObjByIDReply, error) {
	if err := checkGetObjReq(req); err != nil {
		return nil, err
	}

	if len(req.GetIdList()) == 0 {
		return nil, utils.NewStatusError(7000, "find no id")
	}

	reply := &pb.GetObjByIDReply{}

	assemblyObjs := func() error {
		res, err := getObjByIDs(req.GetIdList())
		if err != nil {
			return utils.NewStatusError(7000, err)
		}
		reply.Objs = make(map[int32]*pb.ObjDao)
		for _, ele := range res {
			ti, des := getObjTimeInfoAndDescription(req, ele)
			reply.Objs[int32(ele.ID)] = &pb.ObjDao{
				Id:          int32(ele.ID),
				Name:        ele.Name,
				Description: des,
				TimeInfo:    ti,
			}
		}
		return nil
	}

	assemblyFields := func() error {
		reply.Fields = make(map[int32]*pb.ObjFieldDaoList)
		for _, id := range req.GetIdList() {
			fields, err := getObjFieldsDto(req, int(id))
			if err != nil {
				return utils.NewStatusError(7000, err)
			}
			reply.Fields[id] = &pb.ObjFieldDaoList{
				Objfields: fields,
			}
		}
		return nil
	}

	if req.QueryType == pb.GetObjRequest_OBJ || req.QueryType == pb.GetObjRequest_OBJ_WITH_FIELDS {
		if err := assemblyObjs(); err != nil {
			return nil, err
		}
	}

	if req.QueryType == pb.GetObjRequest_FIELDS || req.QueryType == pb.GetObjRequest_OBJ_WITH_FIELDS {
		if err := assemblyFields(); err != nil {
			return nil, err
		}
	}

	return reply, nil
}

func GetObjByName(ctx context.Context, req *pb.GetObjRequest) (*pb.GetObjByNameReply, error) {
	if err := checkGetObjReq(req); err != nil {
		return nil, err
	}

	if len(req.GetNameList()) == 0 {
		return nil, utils.NewStatusError(7000, "find no name")
	}

	reply := &pb.GetObjByNameReply{}

	objs, err := getObjByNames(req.GetNameList())
	if err != nil {
		return nil, utils.NewStatusError(7000, err)
	}

	assemblyObjs := func() {
		reply.Objs = make(map[string]*pb.ObjDao)
		for _, ele := range objs {
			ti, des := getObjTimeInfoAndDescription(req, ele)
			reply.Objs[ele.Name] = &pb.ObjDao{
				Id:          int32(ele.ID),
				Name:        ele.Name,
				Description: des,
				TimeInfo:    ti,
			}
		}
	}

	assemblyFields := func() error {
		reply.Fields = make(map[string]*pb.ObjFieldDaoList)
		for _, ele := range objs {
			fields, err := getObjFieldsDto(req, int(ele.ID))
			if err != nil {
				return utils.NewStatusError(7000, err)
			}
			reply.Fields[ele.Name] = &pb.ObjFieldDaoList{
				Objfields: fields,
			}
		}
		return nil
	}

	if req.QueryType == pb.GetObjRequest_OBJ || req.QueryType == pb.GetObjRequest_OBJ_WITH_FIELDS {
		assemblyObjs()
	}

	if req.QueryType == pb.GetObjRequest_FIELDS || req.QueryType == pb.GetObjRequest_OBJ_WITH_FIELDS {
		if err := assemblyFields(); err != nil {
			return nil, err
		}
	}

	return reply, nil
}

func checkGetObjReq(req *pb.GetObjRequest) (err error) {
	if req == nil {
		return utils.NewStatusError(7000, "please set request")
	}

	if len(req.IdList) != 0 && len(req.NameList) != 0 {
		return utils.NewStatusError(7000, "you set too many fields")
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

func getObjTimeInfoAndDescription(req *pb.GetObjRequest, obj *model.MtObj) (ti *common.TimeInfo, des *string) {
	if req.GetShowTimeinfo() {
		ti = &common.TimeInfo{
			CreatedAt: obj.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: obj.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	if req.GetShowDescription() {
		des = obj.Description
	}
	return
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

func getFieldTimeInfoAndDescription(req *pb.GetObjRequest, field *model.MtObjField) (ti *common.TimeInfo, des *string) {
	if req.GetShowTimeinfo() {
		ti = &common.TimeInfo{
			CreatedAt: field.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt: field.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}
	if req.GetShowDescription() {
		des = field.Description
	}
	return
}
