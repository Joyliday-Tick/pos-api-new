package models

import (
	"time"
)

type ScoreHistory struct {
	CustomerID       int       `json:"customer_id"`
	ScoreTypeID      int       `json:"score_type_id"`
	ProductID        *int      `json:"product_id"`
	BranchID         int       `json:"branch_id"`
	Amount           int       `json:"amount"`
	Description      *string   `json:"description"`
	TransactionDate  time.Time `json:"transaction_date"`
	CreateBy         string    `json:"create_by"`
	Qty              *int      `json:"qty"`
	MobileNo         string    `json:"mobile_no"`
	RefTransaction   *string   `json:"ref_transaction"`
	RefTransactionID *string   `json:"ref_transaction_id"`
	CardID           *string   `json:"card_id"`
	PosID            *string   `json:"pos_id"`
}

type ScoreMember struct {
	MobileNo   string `json:"mobile_no"`
	Bonus      int    `json:"bonus"`
	TotalPoint int    `json:"total_point"`
	ECoin      int    `json:"e_coin"`
	JubuJibi   int    `json:"jubu_jibi"`
	FinWow     int    `json:"fin_wow"`
	EStamp     int    `json:"e_stamp"`
	MSkill1    int    `json:"m_skill1"`
	MSkill2    int    `json:"m_skill2"`
	MSkill3    int    `json:"m_skill3"`
	MSkill4    int    `json:"m_skill4"`
	MSkill5    int    `json:"m_skill5"`
}
type ScoreType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type BranchCrm struct {
	ID       int    `json:"id"`
	Code     string `json:"code"`
	NameTh   string `json:"name_th"`
	NameEn   string `json:"name_en"`
	IsActive bool   `json:"is_active"`
}

type CustomerCrm struct {
	ID                  int     `json:"id"`
	Firstname           string  `json:"firstname"`
	Lastname            string  `json:"lastname"`
	Email               string  `json:"email"`
	Birthdate           string  `json:"birthdate"` // หรือใช้ time.Time ถ้าจะแปลงเป็นเวลาจริง
	MobileNo            string  `json:"mobile_no"`
	Username            string  `json:"username"`
	IsConsent           bool    `json:"is_consent"`
	IsVerifiedOTP       bool    `json:"is_verified_otp"`
	LineID              string  `json:"line_id"`
	LineUID             string  `json:"line_uid"`
	Sex                 string  `json:"sex"`
	Address1            string  `json:"address1"`
	Address2            string  `json:"address2"`
	DistrictID          string  `json:"district_id"`
	SubdistrictID       string  `json:"subdistrict_id"`
	ProvinceID          string  `json:"province_id"`
	Zipcode             string  `json:"zipcode"`
	Description         string  `json:"description"`
	RoleID              int     `json:"role_id"`
	IsActive            bool    `json:"is_active"`
	CreateBy            *int    `json:"create_by"` // ใช้ pointer เพราะอาจเป็น null
	CreateDate          string  `json:"create_date"`
	UpdateBy            int     `json:"update_by"`
	UpdateDate          string  `json:"update_date"`
	IsDelete            bool    `json:"is_delete"`
	DeleteBy            *int    `json:"delete_by"`
	DeleteDate          *string `json:"delete_date"` // อาจเป็น null
	PictureURL          string  `json:"picture_url"`
	MSkill1             int     `json:"m_skill1"`
	MSkill2             int     `json:"m_skill2"`
	MSkill3             int     `json:"m_skill3"`
	MSkill4             int     `json:"m_skill4"`
	MSkill5             int     `json:"m_skill5"`
	Customer            string  `json:"customer"`
	AppName             string  `json:"app_name"`
	MemberCardNo        string  `json:"member_card_no"`
	MemberClass         string  `json:"member_class"`
	RegisterBy          string  `json:"register_by"`
	RegisterChannel     string  `json:"register_channel"`
	RegisterDate        string  `json:"register_date"`
	Lang                string  `json:"lang"`
	ReferralCode        string  `json:"referral_code"`
	ReferralSource      string  `json:"referral_source"`
	IsFirstTimeRegister string  `json:"is_first_time_register"`
	Bonus               int     `json:"bonus"`
	TotalPoint          int     `json:"total_point"`
	ECoin               int     `json:"e_coin"`
	FinWow              int     `json:"fin_wow"`
	EStamp              int     `json:"e_stamp"`
	CardID              *string `json:"card_id"` // nullable
}

type RowCount struct {
	RowCount int `json:"row_count"`
}

type DeleteRequest struct {
	RefTransaction   string `json:"ref_transaction"`
	RefTransactionID string `json:"ref_transaction_id"`
}

type CustomerCrmResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    CustomerCrm `json:"data"`
}

type BranchCrmResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    BranchCrm `json:"data"`
}

type ScoreTypeCrmResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    []ScoreType `json:"data"`
}

type DeleteHistoryResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    RowCount `json:"data"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		UserID   int    `json:"userId"`
		CustID   int    `json:"custId"`
		Username string `json:"username"`
		Token    string `json:"token"`
		ListMenu string `json:"list_menu"`
		RoleID   int    `json:"role_id"`
		RoleName string `json:"role_name"`
	} `json:"data"`
}
