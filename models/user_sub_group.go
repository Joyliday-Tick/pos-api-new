package models

import (
	"github.com/google/uuid"
)

type UsersSubGroup struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LocationCode string    `gorm:"column:location_code" json:"location_code"`
	GroupID      *int      `gorm:"column:group_id" json:"group_id"`
}

type UsersSubGroupBranch struct {
	BranchList []string `json:"branch_list"`
}

func (UsersSubGroup) TableName() string {
	return "user_sub_group"
}
