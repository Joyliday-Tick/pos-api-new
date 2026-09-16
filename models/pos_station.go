package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type PosStationDto struct {
	PosCode         string   `json:"pos_code"`
	PosDescription  string   `json:"pos_description"`
	PosLocation     string   `json:"pos_location"`
	MacAddress      []string `json:"mac_address"`
	BillHeaderLine1 string   `json:"bill_header_line1"`
	BillHeaderLine2 string   `json:"bill_header_line2"`
	BillHeaderLine3 string   `json:"bill_header_line3"`
	BillHeaderLine4 string   `json:"bill_header_line4"`
	BillHeaderLine5 string   `json:"bill_header_line5"`
	LogoId          string   `json:"logo_id"`
	IsActive        bool     `json:"is_active"`
}

type MacAddressList struct {
	MacAddressList []string `json:"mac_address_list"`
}

type PosStationData struct {
	ID              int             `json:"id"`
	PosCode         string          `json:"pos_code"`
	PosDescription  string          `json:"pos_description"`
	PosLocation     string          `json:"pos_location"`
	MacAddress      json.RawMessage `json:"mac_address"`
	BillHeaderLine1 string          `json:"bill_header_line1"`
	BillHeaderLine2 string          `json:"bill_header_line2"`
	BillHeaderLine3 string          `json:"bill_header_line3"`
	BillHeaderLine4 string          `json:"bill_header_line4"`
	BillHeaderLine5 string          `json:"bill_header_line5"`
	LogoId          *uuid.UUID      `json:"logo_id"`
	LogoName        string          `json:"logo_name"`
	LogoUrl         string          `json:"logo_url"`
	IsActive        bool            `json:"is_active"`
	CreateBy        int             `json:"create_by,omitempty"`
	CreateDate      *time.Time      `json:"create_date"`
	UpdateBy        *int            `json:"update_by"`
	UpdateDate      *time.Time      `json:"update_date"`
	IsDelete        bool            `json:"is_delete"`
	DeleteBy        *int            `json:"delete_by"`
	DeleteDate      *time.Time      `json:"delete_date"`
}

type PosStation struct {
	ID              int             `gorm:"primaryKey;column:id" json:"id"`
	PosCode         string          `gorm:"column:pos_code" json:"pos_code"`
	PosDescription  string          `gorm:"column:pos_description" json:"pos_description"`
	PosLocation     string          `gorm:"column:pos_location" json:"pos_location"`
	MacAddress      json.RawMessage `gorm:"column:mac_address" json:"mac_address"`
	BillHeaderLine1 string          `gorm:"column:bill_header_line1" json:"bill_header_line1"`
	BillHeaderLine2 string          `gorm:"column:bill_header_line2" json:"bill_header_line2"`
	BillHeaderLine3 string          `gorm:"column:bill_header_line3" json:"bill_header_line3"`
	BillHeaderLine4 string          `gorm:"column:bill_header_line4" json:"bill_header_line4"`
	BillHeaderLine5 string          `gorm:"column:bill_header_line5" json:"bill_header_line5"`
	LogoId          *uuid.UUID      `gorm:"column:logo_id" json:"logo_id"`
	IsActive        bool            `gorm:"column:is_active;default:true" json:"is_active"`
	CreateBy        int             `gorm:"column:create_by" json:"create_by,omitempty"`
	CreateDate      *time.Time      `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy        *int            `gorm:"column:update_by" json:"update_by"`
	UpdateDate      *time.Time      `gorm:"column:update_date" json:"update_date"`
	IsDelete        bool            `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy        *int            `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate      *time.Time      `gorm:"column:delete_date" json:"delete_date"`
}

func (PosStation) TableName() string {
	return "pos_station"
}

type SearchStationParams struct {
	Search     string
	MacAddress string
	Location   string
	Page       int
	Skip       int
}
