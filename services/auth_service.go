package services

import (
	"errors"

	"new-pos-api/models"

	"gorm.io/gorm"
)

func Authenticate(username, password string) (*models.AuthResult, error) {
	user, err := FindUserDbByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ชื่อผู้ใช้งานไม่ถูกต้อง")
		}
		return nil, err
	}

	if user == nil {
		return nil, errors.New("ชื่อผู้ใช้งานไม่ถูกต้อง")
	}

	if user.Password != password {
		return nil, errors.New("รหัสผ่านไม่ถูกต้อง")
	}

	// isValid := utils.CheckPasswordHash(password, user.Password)
	// if !isValid {
	// 	return nil, errors.New("รหัสผ่านไม่ถูกต้อง")
	// }

	result := &models.AuthResult{
		Data: user,
		Auth: true,
	}
	return result, nil
}

func AuthenticateWithBranch(username, password string) (*models.AuthResultBranch, error) {
	user, err := FindUserDbByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("ชื่อผู้ใช้งานไม่ถูกต้อง")
		}
		return nil, err
	}

	if user == nil {
		return nil, errors.New("ชื่อผู้ใช้งานไม่ถูกต้อง")
	}

	if user.Password != password {
		return nil, errors.New("รหัสผ่านไม่ถูกต้อง")
	}
	userRole, err := FindUserRoleById(*user.UserRoleId)
	if err != nil {
		return nil, err
	}

	userGroup, err := FindUserGroupById(*user.UserGroupId)
	if err != nil {
		return nil, err
	}

	userSubGroup, err := FindUserSubGroupByGroupId(*user.UserGroupId)
	if err != nil {
		return nil, err
	}

	result := &models.AuthResultBranch{
		Data:       user,
		RoleName:   userRole.RoleDescription,
		GroupName:  userGroup.GroupName,
		BranchList: userSubGroup,
		Auth:       true,
	}
	return result, nil
}
