package model

type MtObjField struct {
	ID          int    `json:"id"`
	ObjID       int    `json:"obj_id"`
	Name        string `json:"name"`
	FieldType   int    `json:"field_type"`
	LookupObjID int    `json:"lookup_obj_id"`
	Description string `json:"description"`
	IsAllowNull bool   `json:"is_allow_null"`
	*TimeInfo
}

func (MtObjField) TableName() string {
	return "mt_obj_field"
}
