package models

import (
	"time"

	"github.com/google/uuid"
)

type AdjustPointDto struct {
	Username     string    `json:"username"`
	MemberTel    string    `json:"member_tel"`
	Reason       string    `json:"reason"`
	AdjBranch    string    `json:"branch"`
	AdjPoint     int       `json:"adj_point"`
	AdjJoylicoin int       `json:"adj_joylicoin"`
	AdjFinwow    int       `json:"adj_finwow"`
	AdjEstamp    int       `json:"adj_estamp"`
	AdjMskill1   int       `json:"adj_power"`
	AdjMskill2   int       `json:"adj_agility"`
	AdjMskill3   int       `json:"adj_reaction_time"`
	AdjMskill4   int       `json:"adj_balance"`
	AdjMskill5   int       `json:"adj_speed"`
	AdjJubuJibi  int       `json:"adj_jubu_jibi"`
	AdjDate      time.Time `json:"-"`
}

type AdjustPoint struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey;column:id" json:"id"`
	Username     string    `gorm:"column:username" json:"username"`
	MemberTel    string    `gorm:"column:member_tel" json:"member_tel"`
	Reason       string    `gorm:"column:reason" json:"reason"`
	AdjDate      time.Time `gorm:"column:adj_date;default:now()" json:"adj_date"`
	AdjBranch    string    `gorm:"column:adj_branch" json:"adj_branch"`
	AdjPoint     int       `gorm:"default:0;column:adj_point" json:"adj_point"`
	AdjJoylicoin int       `gorm:"default:0;column:adj_joylicoin" json:"adj_joylicoin"`
	AdjFinwow    int       `gorm:"default:0;column:adj_finwow" json:"adj_finwow"`
	AdjEstamp    int       `gorm:"default:0;column:adj_estamp" json:"adj_estamp"`
	AdjMskill1   int       `gorm:"default:0;column:adj_mskill1" json:"adj_mskill1"`
	AdjMskill2   int       `gorm:"default:0;column:adj_mskill2" json:"adj_mskill2"`
	AdjMskill3   int       `gorm:"default:0;column:adj_mskill3" json:"adj_mskill3"`
	AdjMskill4   int       `gorm:"default:0;column:adj_mskill4" json:"adj_mskill4"`
	AdjMskill5   int       `gorm:"default:0;column:adj_mskill5" json:"adj_mskill5"`
	AdjJubuJibi  int       `gorm:"default:0;column:adj_jubu_jibi" json:"adj_jubu_jibi"`
}

type AdjustPointData struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	MemberTel    string    `json:"member_tel"`
	Reason       string    `json:"reason"`
	AdjDate      time.Time `json:"adj_date"`
	AdjBranch    string    `json:"adj_branch"`
	AdjPoint     int       `json:"adj_point"`
	AdjJoylicoin int       `json:"adj_joylicoin"`
	AdjFinwow    int       `json:"adj_finwow"`
	AdjEstamp    int       `json:"adj_estamp"`
	AdjMskill1   int       `json:"adj_mskill1"`
	AdjMskill2   int       `json:"adj_mskill2"`
	AdjMskill3   int       `json:"adj_mskill3"`
	AdjMskill4   int       `json:"adj_mskill4"`
	AdjMskill5   int       `json:"adj_mskill5"`
	AdjJubuJibi  int       `json:"adj_jubu_jibi"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
}

type AdjustPointDataExport struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	MemberTel    string    `json:"member_tel"`
	Reason       string    `json:"reason"`
	AdjDate      time.Time `json:"adj_date"`
	AdjBranch    string    `json:"adj_branch"`
	AdjPoint     int       `json:"adj_point"`
	AdjJoylicoin int       `json:"adj_joylicoin"`
	AdjFinwow    int       `json:"adj_finwow"`
	AdjEstamp    int       `json:"adj_estamp"`
	AdjMskill1   int       `json:"adj_mskill1"`
	AdjMskill2   int       `json:"adj_mskill2"`
	AdjMskill3   int       `json:"adj_mskill3"`
	AdjMskill4   int       `json:"adj_mskill4"`
	AdjMskill5   int       `json:"adj_mskill5"`
	AdjJubuJibi  int       `json:"adj_jubu_jibi"`
	Firstname    string    `json:"firstname"`
	Lastname     string    `json:"lastname"`
	MemberId     int       `json:"member_id"`
}

func (AdjustPoint) TableName() string {
	return "adjust_point"
}

type SearchAdjustPointReport struct {
	StartDate string
	EndDate   string
	Location  string
	MemberTel string
	Page      int
	Skip      int
}

type ExportAdjustPointReport struct {
	StartDate    string
	EndDate      string
	Location     string
	MemberTel    string
	ExportFormat string // "pdf" or "excel"
}
