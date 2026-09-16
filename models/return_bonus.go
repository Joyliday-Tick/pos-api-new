package models

import (
	"time"

	"github.com/google/uuid"
)

type ReturnBonusDto struct {
	BonusNo      string    `json:"-"`
	BonusDate    time.Time `json:"-"`
	MemberTel    string    `json:"member_tel"`
	Branch       string    `json:"branch"`
	SetBonus     *float64  `json:"persent"`
	SetPointRate *float64  `json:"price_to_point"`
	CSpend       *int64    `json:"total_spend"`
	PointStamp   *int64    `json:"final_point"`
	// PointScan     *int64    `json:"point_scan"`
	BonusReturn *int64 `json:"joylicoin_return"`
	Employee    string `json:"cashier"`
	// PointAdj      int     `json:"point_adj"`
	// BonusRedeem   int     `json:"bonus_redeem"`
	// Discount      float64 `json:"discount"`
	LastPoint     int    `json:"member_point"`
	LastJoylicoin int    `json:"member_joylicoin"`
	Extra         int    `json:"extra"`
	EstampBaht    int    `json:"estamp_baht"`
	LastEstamp    int    `json:"member_ecoin"`
	PosId         string `json:"pos_id"`
}

type ReturnBonus struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BonusNo       string    `json:"bonus_no" gorm:"column:bonus_no"`
	BonusDate     time.Time `json:"bonus_date" gorm:"column:bonus_date"`
	MemberTel     string    `json:"member_tel" gorm:"column:member_tel"`
	Branch        string    `json:"branch" gorm:"column:branch"`
	SetBonus      *float64  `json:"set_bonus"  gorm:"column:set_bonus"`
	SetPointRate  *float64  `json:"set_point_rate"  gorm:"column:set_point_rate"`
	CSpend        *int64    `json:"c_spend" gorm:"column:c_spend"`
	PointStamp    *int64    `json:"point_stamp"  gorm:"column:point_stamp"`
	PointScan     *int64    `json:"point_scan"  gorm:"column:point_scan;default:0"`
	BonusReturn   *int64    `json:"bonus_return"  gorm:"column:bonus_return"`
	Employee      string    `json:"employee"  gorm:"column:employee"`
	IntSpend      int       `json:"int_spend"  gorm:"column:int_spend;default:0"`
	PointAdj      int       `json:"point_adj"  gorm:"column:point_adj;default:0"`
	BonusRedeem   int       `json:"bonus_redeem"  gorm:"column:bonus_redeem;default:0"`
	Discount      float64   `json:"discount"  gorm:"column:discount;default:0.00"`
	LastPoint     int       `json:"last_point"  gorm:"column:last_point;default:0"`
	LastJoylicoin int       `json:"last_joylicoin"  gorm:"column:last_joylicoin;default:0"`
	Extra         int       `json:"extra"  gorm:"column:extra;default:0"`
	EstampBaht    int       `json:"estamp_baht"  gorm:"column:estamp_baht;default:0"`
	LastEstamp    int       `json:"last_estamp"  gorm:"column:last_estamp;default:0"`
}

func (ReturnBonus) TableName() string {
	return "return_bonus"
}

type SearchClaimReport struct {
	StartDate string
	EndDate   string
	Location  []string
	MemberTel string
	Page      int
	Skip      int
}

type ExportClaimReport struct {
	StartDate string
	EndDate   string
	Location  []string
	MemberTel string
}

type SearchClaimDetail struct {
	StartDate string
	EndDate   string
	// Location  string
	MemberTel string
}

type ReturnBonusData struct {
	ID            uuid.UUID `json:"id"`
	BonusNo       string    `json:"bonus_no"`
	BonusDate     time.Time `json:"bonus_date"`
	MemberTel     string    `json:"member_tel"`
	Branch        string    `json:"branch"`
	CSpend        *int64    `json:"total_spend"`
	PointStamp    *int64    `json:"point_stamp"`
	BonusReturn   *int64    `json:"joylicoin"`
	Employee      string    `json:"employee"`
	LastPoint     int       `json:"last_point"`
	LastJoylicoin int       `json:"last_joylicoin"`
	Extra         int       `json:"extra"`
	EstampBaht    int       `json:"estamp_baht"`
	LastEstamp    int       `json:"last_estamp"`
}

type ReturnBonusDataExport struct {
	ID            uuid.UUID `json:"id"`
	BonusNo       string    `json:"bonus_no"`
	BonusDate     time.Time `json:"bonus_date"`
	MemberTel     string    `json:"member_tel"`
	Branch        string    `json:"branch"`
	CSpend        *int64    `json:"total_spend"`
	PointStamp    *int64    `json:"point_stamp"`
	BonusReturn   *int64    `json:"joylicoin"`
	Employee      string    `json:"employee"`
	LastPoint     int       `json:"last_point"`
	LastJoylicoin int       `json:"last_joylicoin"`
	Extra         int       `json:"extra"`
	EstampBaht    int       `json:"estamp_baht"`
	LastEstamp    int       `json:"last_estamp"`
	MemberId      int       `json:"member_id"`
}

type ClaimData struct {
	ClaimList      []ReturnBonusData `json:"claim_list"`
	TotalJoylicoin int64             `json:"total_joylicoin"`
	TotalSpending  int64             `json:"total_spending"`
	Percent        float64           `json:"percent"`
}

type ClaimDetail struct {
	MachineNo   int    `json:"machine_no"`
	AssetCode   string `json:"machine_asset"`
	MachineName string `json:"machine_name"`
	// AssetHead    int    `json:"asset_head"`
	Location string `json:"location"`
	ECoin    int    `json:"used_ecoin"`
	EBonus   int    `json:"used_bonus"`
}
