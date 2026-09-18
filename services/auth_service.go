package services

import (
	"errors"
	"fmt"

	"new-pos-api/models"

	"gorm.io/gorm"
)

// ห้ามชื่อผู้ใช้หรือรหัสผ่านว่าง ไม่ว่าจะมาทางไหน
//
// ⚠️ นี่ไม่ใช่การกันไว้เฉย ๆ — เมื่อ 18 ก.ย. ยิงจริงแล้วเข้าได้
//
//	curl -H "Authorization: Basic Og==" .../api/user-role   -> 200
//
// "Og==" คือ base64 ของ ":" ซึ่งถอดออกมาได้ชื่อผู้ใช้ว่างกับรหัสผ่านว่าง
// ในตาราง users มีแถวหนึ่ง (id 24) ที่ทั้ง username และ password เป็นค่าว่าง
// และยัง is_active = true อยู่ FindUserDbByUsername จึงหาเจอ แล้ว
// การเทียบ user.Password != password กลายเป็น "" != "" ซึ่งเป็นเท็จ = ผ่าน
//
// controllers.Login มีด่านกันค่าว่างอยู่ แต่เส้นทาง Basic auth ใน
// middlewares/auth.go เรียก Authenticate ตรง ๆ โดยไม่ผ่านด่านนั้น
// และ Basic auth ใช้ได้กับทุก endpoint จึงเปิดทั้ง API ด้วย header บรรทัดเดียว
//
// วางด่านไว้ที่ชั้น service เพราะเป็นจุดเดียวที่ผู้เรียกทุกทางต้องผ่าน
// การไปแก้เฉพาะ controller หรือเฉพาะ middleware จะพลาดอีกทางเสมอ
func rejectEmptyCredentials(username, password string) error {
	if username == "" {
		return errors.New("ชื่อผู้ใช้งานไม่ถูกต้อง")
	}
	if password == "" {
		return errors.New("รหัสผ่านไม่ถูกต้อง")
	}
	return nil
}

func Authenticate(username, password string) (*models.AuthResult, error) {
	if err := rejectEmptyCredentials(username, password); err != nil {
		return nil, err
	}

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
	if err := rejectEmptyCredentials(username, password); err != nil {
		return nil, err
	}

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
	// คอลัมน์ user_role_id และ user_group_id เป็น nullable ทั้งคู่ และ
	// FindUserRoleById/FindUserGroupById คืน (nil, nil) เมื่อหาไม่เจอ
	// ของเดิม deref ทั้งตัว pointer และผลลัพธ์โดยไม่ตรวจ รวม 5 ทางที่ panic ได้
	// gin.Recovery() รับไว้เป็น 500 โปรเซสไม่ตาย แต่คนล็อกอินเห็นแค่
	// "เกิดข้อผิดพลาด" โดยไม่มีอะไรบอกว่าปัญหาอยู่ที่ข้อมูลผู้ใช้คนนั้น
	//
	// ข้อมูล UAT ปัจจุบันสะอาดทั้ง 4 ทาง (ตรวจ 18 ก.ย. ไม่มีแถวไหนเข้าเงื่อนไข)
	// นี่จึงเป็นการกันไว้ ไม่ใช่การแก้อาการที่เกิดอยู่
	// ทางที่เข้าถึงง่ายที่สุดคือปิดกลุ่มผู้ใช้ (is_active = false) ทั้งที่ยังมีคนอยู่ในกลุ่ม
	// เพราะ FindUserGroupById กรอง is_active = true ด้วย
	if user.UserRoleId == nil {
		return nil, errors.New("บัญชีนี้ยังไม่ได้กำหนดสิทธิ์การใช้งาน กรุณาติดต่อผู้ดูแลระบบ")
	}
	if user.UserGroupId == nil {
		return nil, errors.New("บัญชีนี้ยังไม่ได้กำหนดกลุ่มผู้ใช้ กรุณาติดต่อผู้ดูแลระบบ")
	}

	userRole, err := FindUserRoleById(*user.UserRoleId)
	if err != nil {
		return nil, err
	}
	if userRole == nil {
		return nil, fmt.Errorf(
			"ไม่พบสิทธิ์การใช้งานรหัส %d ที่ผูกกับบัญชีนี้ (อาจถูกลบไปแล้ว) กรุณาติดต่อผู้ดูแลระบบ",
			*user.UserRoleId)
	}

	userGroup, err := FindUserGroupById(*user.UserGroupId)
	if err != nil {
		return nil, err
	}
	if userGroup == nil {
		return nil, fmt.Errorf(
			"ไม่พบกลุ่มผู้ใช้รหัส %d ที่ผูกกับบัญชีนี้ (อาจถูกลบหรือถูกปิดใช้งาน) กรุณาติดต่อผู้ดูแลระบบ",
			*user.UserGroupId)
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
