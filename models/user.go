package models

import "time"

type UsersCreateDto struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	SName       string `json:"s_name"`
	Status      string `json:"status"`
	UserRoleId  *int   `json:"user_role_id"`
	UserGroupId *int   `json:"user_group_id"`
	IsActive    bool   `json:"is_active"`
}

type UsersUpdateDto struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	NewPassword string `json:"new_password"`
	Name        string `json:"name"`
	SName       string `json:"s_name"`
	Status      string `json:"status"`
	UserRoleId  *int   `json:"user_role_id"`
	UserGroupId *int   `json:"user_group_id"`
	IsActive    bool   `json:"is_active"`
}

type UsersData struct {
	ID            int        `json:"id"`
	Username      string     `json:"username"`
	Password      string     `json:"password"`
	Name          string     `json:"name"`
	SName         string     `json:"s_name"`
	Status        string     `json:"status"`
	UserRoleId    *int       `json:"user_role_id"`
	UserGroupId   *int       `json:"user_group_id"`
	IsActive      bool       `json:"is_active"`
	CreateBy      int        `json:"create_by"`
	CreateDate    *time.Time `json:"create_date"`
	UpdateBy      *int       `json:"update_by"`
	UpdateDate    *time.Time `json:"update_date"`
	IsDelete      bool       `json:"is_delete"`
	DeleteBy      *int       `json:"delete_by"`
	DeleteDate    *time.Time `json:"delete_date"`
	UserRoleName  string     `json:"user_role_name"`
	UserGroupName string     `json:"user_group_name"`
}

type Users struct {
	ID          int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username    string     `gorm:"column:username" json:"username"`
	Password    string     `gorm:"column:password" json:"password"`
	Name        string     `gorm:"column:name" json:"name"`
	SName       string     `gorm:"column:s_name" json:"s_name"`
	Status      string     `gorm:"column:status" json:"status"`
	UserRoleId  *int       `gorm:"column:user_role_id" json:"user_role_id"`
	UserGroupId *int       `gorm:"column:user_group_id" json:"user_group_id"`
	IsActive    bool       `gorm:"column:is_active" json:"is_active"`
	CreateBy    int        `gorm:"column:create_by" json:"create_by"`
	CreateDate  *time.Time `gorm:"column:create_date;default:now()" json:"create_date"`
	UpdateBy    *int       `gorm:"column:update_by" json:"update_by"`
	UpdateDate  *time.Time `gorm:"column:update_date" json:"update_date"`
	IsDelete    bool       `gorm:"column:is_delete;default:false" json:"is_delete"`
	DeleteBy    *int       `gorm:"column:delete_by" json:"delete_by"`
	DeleteDate  *time.Time `gorm:"column:delete_date" json:"delete_date"`
}

func (Users) TableName() string {
	return "users"
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResult struct {
	Data *Users `json:"data"`
	Auth bool   `json:"auth"`
}

type AuthResultBranch struct {
	Data       *Users   `json:"data"`
	Auth       bool     `json:"auth"`
	RoleName   string   `json:"role_name"`
	GroupName  string   `json:"group_name"`
	BranchList []string `json:"branch_list"`
}

type SearchUserParams struct {
	Search      string
	UserRoleId  *int
	UserGroupId *int
	Username    string
	Page        int
	Skip        int
}

type UserLite struct {
	ID    int
	Name  string
	SName string
}
