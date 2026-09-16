package models

import (
	"time"

	"github.com/google/uuid"
)

type CardPlayBranchDto struct {
	CardPlayId uuid.UUID `json:"card_play_id"`
	BranchCode string    `json:"branch_code"`
}


type CardPlayBranch struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CardPlayId uuid.UUID  `json:"card_play_id"`
	BranchCode string     `json:"branch_code"`
	CreateBy   int        `json:"create_by"`
	CreateDate time.Time  `json:"create_date" gorm:"default:CURRENT_TIMESTAMP"`
	UpdateBy   *int       `json:"update_by"`
	UpdateDate *time.Time `json:"update_date"`
	IsDelete   bool       `json:"is_delete" gorm:"default:false"`
	DeleteBy   *int       `json:"delete_by"`
	DeleteDate *time.Time `json:"delete_date"`
}

func (CardPlayBranch) TableName() string {
	return "card_play_branch"
}
