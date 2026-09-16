package models

import (
	"time"
)

type MemberDto struct {
	Tel        string `json:"tel"`
	MName      string `json:"m_name"`
	SName      string `json:"s_name"`
	TotalPoint int    `json:"total_point"`
	Bonus      int    `json:"bonus"`
	Ecoin      int    `json:"ecoin"`
	Finwow     int    `json:"finwow"`
	Estamp     int    `json:"estamp"`
	MSkill1    int    `json:"mskill1"`
	MSkill2    int    `json:"mskill2"`
	MSkill3    int    `json:"mskill3"`
	MSkill4    int    `json:"mskill4"`
	MSkill5    int    `json:"mskill5"`
	JubuJibi   int    `json:"jubu_jibi"`
	IsActive   bool   `json:"is_active"`
}

type MemberNameDto struct {
	Tel   string `json:"tel"`
	MName string `json:"m_name"`
	SName string `json:"s_name"`
}

type SkillDto struct {
	MSkill1 int `json:"mskill1"`
	MSkill2 int `json:"mskill2"`
	MSkill3 int `json:"mskill3"`
	MSkill4 int `json:"mskill4"`
	MSkill5 int `json:"mskill5"`
}

type JoylicoinDto struct {
	Joylicoin int `json:"joylicoin"`
}

type TotalPointDto struct {
	TotalPoint int `json:"total_point"`
}

type Member struct {
	ID         int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Tel        string     `gorm:"column:tel;type:varchar(10)" json:"tel"`
	MName      string     `gorm:"column:m_name;type:varchar(100)" json:"m_name"`
	SName      string     `gorm:"column:s_name;type:varchar(100)" json:"s_name"`
	TotalPoint int        `gorm:"column:total_point;default:0" json:"total_point"`
	Bonus      int        `gorm:"column:bonus;default:0" json:"bonus"`
	Ecoin      int        `gorm:"column:ecoin;default:0" json:"ecoin"`
	Finwow     int        `gorm:"column:finwow;default:0" json:"finwow"`
	Estamp     int        `gorm:"column:estamp;default:0" json:"estamp"`
	MSkill1    int        `gorm:"column:mskill1;default:0" json:"mskill1"`
	MSkill2    int        `gorm:"column:mskill2;default:0" json:"mskill2"`
	MSkill3    int        `gorm:"column:mskill3;default:0" json:"mskill3"`
	MSkill4    int        `gorm:"column:mskill4;default:0" json:"mskill4"`
	MSkill5    int        `gorm:"column:mskill5;default:0" json:"mskill5"`
	JubuJibi   int        `gorm:"column:jubu_jibi;default:0" json:"jubu_jibi"`
	IsActive   bool       `gorm:"column:is_active;default:true" json:"is_active"`
	CreateDate time.Time  `gorm:"column:create_date;autoCreateTime" json:"create_date"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"update_date"`
	IsDelete   bool       `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteDate *time.Time `gorm:"column:delete_date" json:"delete_date"`
}

// TableName overrides the table name used by GORM
func (Member) TableName() string {
	return "member" // lowercase + quoted table name
}

type MemberWithTier struct {
	Member
	Tier *string `json:"tier"`
}

type EStampMember struct {
	ID         int     `gorm:"column:ID;primaryKey;autoIncrement"`
	Tel        *string `gorm:"column:Tel;size:10"`
	Mname      *string `gorm:"column:Mname;size:100"`
	Sname      *string `gorm:"column:Sname;size:100"`
	TotalPoint int     `gorm:"column:TotalPoint;default:0"`
	Bonus      int     `gorm:"column:Bonus;default:0"`
	IsDeleted  int8    `gorm:"column:IsDeleted;default:0"`
	Ecoin      int     `gorm:"column:ecoin;default:0"`
	Finwow     int     `gorm:"column:finwow;default:0"`
	Estamp     int     `gorm:"column:estamp;default:0"`
	MSkill1    int     `gorm:"column:mSkill1;default:0"`
	MSkill2    int     `gorm:"column:mSkill2;default:0"`
	MSkill3    int     `gorm:"column:mSkill3;default:0"`
	MSkill4    int     `gorm:"column:mSkill4;default:0"`
	MSkill5    int     `gorm:"column:mSkill5;default:0"`
	JubuJibi   int     `gorm:"column:jubu_jibi;default:0"`
}
