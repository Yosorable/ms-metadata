package model

type MtObjIndex struct {
	ID          int    `json:"id"`
	ObjFieldIds string `json:"obj_field_ids"`
	IndexType   int    `json:"index_type"`
	*TimeInfo
}

func (MtObjIndex) TableName() string {
	return "mt_obj_index"
}
