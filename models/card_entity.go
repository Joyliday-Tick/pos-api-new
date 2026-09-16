package models

import (
	"database/sql"
	"time"

	// "encoding/json"

	"github.com/google/uuid"
)

type CardRegisterDto struct {
	CardNo      string `json:"card_no"`
	CardTypeId  string `json:"card_type_id"`
	IsActive    bool   `json:"is_active"`
	MemberTel   string `json:"member_tel"`
	ECoin       int    `json:"e_coin"`
	EBonus      int    `json:"e_bonus"`
	FromChannel string `json:"from_channel"`
}

type CardSummaryDto struct {
	CardNo       string  `json:"card_no"`
	CardType     string  `json:"card_type"`
	ShowBalance  string  `json:"show_balance"`
	MemberTel    string  `json:"member_tel"`
	ECoin        int     `json:"e_coin"`
	EBonus       int     `json:"e_bonus"`
	DiscountCash float64 `json:"discount_cash"`
	CustomerTier string  `json:"customer_tier"`
}

type CardListData struct {
	MobileNo    string  `json:"mobile_no"`
	CardNo      string  `json:"card_no"`
	CardType    string  `json:"card_type"`
	ECoin       int     `json:"balance_e_coin"`
	EBonus      int     `json:"balance_e_bonus"`
	TopupAmount float64 `json:"topup_amount"`
}

type PosTopupDto struct {
	CardNo         string     `json:"card_no" validate:"required"`
	CardTypeId     string     `json:"card_type_id" validate:"required"`
	Products       []Products `json:"products" validate:"required"`
	IsActive       bool       `json:"is_active" validate:"required"`
	MemberTel      string     `json:"member_tel" validate:"required,len=10,numeric"`
	Cashier        string     `json:"cachier" validate:"required"`
	PosId          string     `json:"pos_id" validate:"required"`
	BillLocation   string     `json:"bill_location" validate:"required"`
	BillPaymenytId string     `json:"bill_payment_id" validate:"required"`
	PosType        string     `json:"pos_type" validate:"required"`
	BankDetail     *string    `json:"bank_detail"` // <-- nullable string
	FreePoint      *int       `json:"free_point"`
	PosMenuId      int        `json:"pos_menu_id" validate:"required"`
	FromChannel    string     `json:"from_channel"`
	BillNo         *string    `json:"bill_no"`
}

type CardEntity struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CardNo     string     `json:"card_no"`
	CardTypeId uuid.UUID  `json:"card_type_id"`
	MemberTel  string     `json:"member_tel"`
	IsActive   bool       `json:"is_active" gorm:"default:true"`
	CreateBy   int        `json:"create_by"`
	CreateDate time.Time  `json:"create_date" gorm:"default:CURRENT_TIMESTAMP"`
	UpdateBy   *int       `json:"update_by"`
	UpdateDate *time.Time `json:"update_date"`
	IsDelete   bool       `json:"is_delete" gorm:"default:false"`
	DeleteBy   *int       `json:"delete_by"`
	DeleteDate *time.Time `json:"delete_date"`
}

// products struct
// productid
// quantity
type Products struct {
	ProductId int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func (CardEntity) TableName() string {
	return "card_entity"
}

type ClearCardDto struct {
	CardNo        string `gorm:"type:varchar(15)" json:"card_no"`
	MemberTel     string `gorm:"type:varchar(20)" json:"member_tel"`
	BalanceEcoin  int    `gorm:"type:integer" json:"balance_ecoin"`
	BalanceEbonus int    `gorm:"type:integer" json:"balance_ebonus"`
	CreatedBy     string `gorm:"type:varchar(45)" json:"created_by"`
	Location      string `gorm:"type:varchar(10)" json:"location"`
	TimePlay      int    `gorm:"type:integer; default:0" json:"time_play"`
	CardType      string `gorm:"type:varchar(50)" json:"card_type"`
}

type ClearCard struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ClearDate     time.Time `gorm:"type:timestamp;not null;default:now()" json:"clear_date"`
	CardNo        string    `gorm:"type:varchar(15)" json:"card_no"`
	MemberTel     string    `gorm:"type:varchar(20)" json:"member_tel"`
	BalanceEcoin  int       `gorm:"type:integer" json:"balance_ecoin"`
	BalanceEbonus int       `gorm:"type:integer" json:"balance_ebonus"`
	CreatedBy     string    `gorm:"type:varchar(45)" json:"created_by"`
	Location      string    `gorm:"type:varchar(10)" json:"location"`
	TimePlay      int       `gorm:"type:integer; default:0" json:"time_play"`
	CardType      string    `gorm:"type:varchar(50)" json:"card_type"`
}

func (ClearCard) TableName() string {
	return "clear_card"
}

type CardDetailDto struct {
	CardNo        string     `json:"card_no"`
	CardType      string     `json:"card_type"`
	ShowBalance   string     `json:"show_balance"`
	MemberTel     string     `json:"member_tel"`
	ECoin         int        `json:"e_coin"`
	EBonus        int        `json:"e_bonus"`
	EBonusExpDate *time.Time `json:"e_bonus_exp_date"`
	EBonusExpDays string     `json:"e_bonus_exp_days"`
	ETimes        int        `json:"e_times"`
	CardExpDate   *time.Time `json:"card_exp_date"`
	CardExpDays   string     `json:"card_exp_days"`
	DiscountCash  float64    `json:"discount_cash"`
	IsMachine     bool       `json:"is_machine"`
	CustomerTier  string     `json:"customer_tier"`
}

type BonusExpireDto struct {
	EBonusExpDate time.Time `json:"e_bonus_exp_date,omitempty"`
	EBonusExpDays string    `json:"e_bonus_exp_days"`
}

type EtimesDto struct {
	PlayTime int `json:"play_time"`
	UsedTime int `json:"used_time"`
}
type CardExpireDto struct {
	CardExpDate    *time.Time   `json:"card_exp_date,omitempty"`
	CardExpDays    string       `json:"card_exp_days"`
	CardExpireDate sql.NullTime `json:"card_expire_date"`
}

type ClearCardData struct {
	ClearCardList []ClearCard `json:"clear_card_list"`
	TotalECoin    int64       `json:"total_ecoin"`
	TotalEBonus   int64       `json:"total_ebonus"`
	TotalTimePlay int64       `json:"total_time_play"`
}

type ExportClearCardReport struct {
	StartDate string
	EndDate   string
	Location  string
}

type CardPackageDto struct {
	CardNo       string `json:"card_no"`
	CardType     string `json:"card_type"`
	MemberTel    string `json:"member_tel"`
	CustomerTier string `json:"customer_tier"`
}

type CardPackageDetailDto struct {
	MachineId      int    `json:"machine_id"`
	MachineName    string `json:"machine_name"`
	ECoin          int    `json:"e_coin"`
	EBonus         int    `json:"e_bonus"`
	PlayTime       int    `json:"play_time"`
	RemainECoin    int    `json:"remain_e_coin"`
	RemainEBonus   int    `json:"remain_e_bonus"`
	RemainPlayTime int    `json:"remain_play_time"`
	UsedECoin      int    `json:"used_e_coin"`
	UsedEBonus     int    `json:"used_e_bonus"`
	UsedPlayTime   int    `json:"used_play_time"`
}

type CardPackageDetailV1Dto struct {
	MachineId      int    `json:"machine_id"`
	MachineName    string `json:"machine_name"`
	Category       string `json:"category"`
	ECoin          int    `json:"e_coin"`
	EBonus         int    `json:"e_bonus"`
	PlayTime       int    `json:"play_time"`
	RemainECoin    int    `json:"remain_e_coin"`
	RemainEBonus   int    `json:"remain_e_bonus"`
	RemainPlayTime int    `json:"remain_play_time"`
}

type CardPackageSummaryDto struct {
	CardNo       string                 `json:"card_no"`
	CardType     string                 `json:"card_type"`
	MemberTel    string                 `json:"member_tel"`
	CustomerTier string                 `json:"customer_tier"`
	PosMenuName  string                 `json:"pos_menu_name"`
	CardExpDate  *time.Time             `json:"card_exp_date"`
	CardExpDays  string                 `json:"card_exp_days"`
	CardPackages []CardPackageDetailDto `json:"card_packages"`
}

type CardEntityType struct {
	CardTypeName string `json:"card_type_name"`
}

type CardPackageAccDto struct {
	ECoin    int `json:"e_coin"`
	EBonus   int `json:"e_bonus"`
	PlayTime int `json:"play_time"`
}

type ClearCardExport struct {
	ID            uuid.UUID `json:"id"`
	ClearDate     time.Time `json:"clear_date"`
	CardNo        string    `json:"card_no"`
	MemberTel     string    `json:"member_tel"`
	BalanceEcoin  int       `json:"balance_ecoin"`
	BalanceEbonus int       `json:"balance_ebonus"`
	CreatedBy     string    `json:"created_by"`
	Location      string    `json:"location"`
	TimePlay      int       `json:"time_play"`
	CardType      string    `json:"card_type"`
	MemberID      *int      `json:"member_id"`
}
