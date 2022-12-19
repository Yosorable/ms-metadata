package model

type MtObj struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	*TimeInfo
}

func (MtObj) TableName() string {
	return "mt_obj"
}
