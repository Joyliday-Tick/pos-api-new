package models

type MasterActionDto struct {
	ID         int    `json:"id"`
	ActionName string `json:"action_name"`
	IsActive   bool   `json:"is_active"`
	IsDelete   bool   `json:"-"`
}

type MasterAction struct {
	ID         int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ActionName string `gorm:"column:action_name;type:varchar(100);not null" json:"action_name"`
	IsActive   bool   `gorm:"column:is_active;default:true;not null" json:"is_active"`
	IsDelete   bool   `gorm:"column:is_delete;default:false;not null" json:"is_delete"`
}

func (MasterAction) TableName() string {
	return "master_action"
}
