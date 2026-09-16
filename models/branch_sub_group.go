package models

import "github.com/google/uuid"

type PosBranchSubGroup struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	LocationCode string    `gorm:"column:location_code" json:"location_code"`
	GroupID      *int      `gorm:"column:group_id" json:"group_id"`
}

func (PosBranchSubGroup) TableName() string {
	return "pos_branch_sub_group"
}
