package utils

import (
	"testing"

	"new-pos-api/models"

	"github.com/golang-jwt/jwt"
)

// ไฟล์ทดสอบไฟล์แรกของโปรเจกต์ เขียนไว้เพราะข้อนี้แก้แล้วดูด้วยตาไม่ออก
// ความต่างระหว่าง "ถูก" กับ "ผิด" คือ claims["roleId"] เป็นตัวเลข 0
// กับเป็น JSON null ซึ่งทั้งสองแบบทำให้ล็อกอินผ่านเหมือนกัน
// ต่างกันตรงที่แบบหลังทำให้ทุกหน้าที่มี RequireRole ตอบ 403 แบบไม่บอกสาเหตุ
//
// ทดสอบล้วน ๆ ไม่แตะฐานข้อมูล — GetJWT ไม่ได้ query อะไร

func withJwtEnv(t *testing.T, tokenExpire string) {
	t.Helper()
	t.Setenv("JWT_SECRET", "กุญแจทดสอบยาวพอสำหรับ HS256 อย่างน้อยสามสิบสองไบต์")
	t.Setenv("TOKEN_EXPIRE", tokenExpire)
	if err := InitJwtSecret(); err != nil {
		t.Fatalf("InitJwtSecret: %v", err)
	}
	if err := InitTokenExpire(); err != nil {
		t.Fatalf("InitTokenExpire: %v", err)
	}
}

func claimsOf(t *testing.T, tokenStr string) jwt.MapClaims {
	t.Helper()
	claims, err := VerifyJwt(tokenStr)
	if err != nil {
		t.Fatalf("VerifyJwt: %v", err)
	}
	return claims
}

// role ว่าง (คอลัมน์ user_role_id เป็น NULL) ต้องได้ตัวเลข 0 ไม่ใช่ null
func TestGetJWT_NilRoleBecomesZeroNotNull(t *testing.T) {
	withJwtEnv(t, "1")

	data, err := GetJWT(models.AuthResultBranch{
		Data: &models.Users{ID: 999, Username: "ทดสอบ", UserRoleId: nil},
	}, "test")
	if err != nil {
		t.Fatalf("GetJWT: %v", err)
	}

	if got := data["roleId"]; got != 0 {
		t.Errorf("response roleId = %#v ต้องเป็น 0", got)
	}

	raw := claimsOf(t, data["token"].(string))["roleId"]
	if raw == nil {
		t.Fatal("claims roleId เป็น null — GetRoleIdFromClaims จะคืน error " +
			"แล้ว RequireRole จะตอบ 403 โดยไม่บอกว่าสาเหตุคือ role ว่าง")
	}
	// ผ่าน JSON แล้วตัวเลขกลายเป็น float64 เสมอ ซึ่ง GetRoleIdFromClaims รับได้
	if v, ok := raw.(float64); !ok || v != 0 {
		t.Errorf("claims roleId = %#v (%T) ต้องเป็น 0", raw, raw)
	}
}

// role ปกติต้องส่งต่อมาครบ ไม่ถูกกลบเป็น 0 ไปด้วย
func TestGetJWT_NormalRolePassesThrough(t *testing.T) {
	withJwtEnv(t, "1")

	role := 3
	data, err := GetJWT(models.AuthResultBranch{
		Data: &models.Users{ID: 42, Username: "ผู้จัดการ", UserRoleId: &role},
	}, "test")
	if err != nil {
		t.Fatalf("GetJWT: %v", err)
	}

	if got := data["roleId"]; got != 3 {
		t.Errorf("response roleId = %#v ต้องเป็น 3", got)
	}
	if v, ok := claimsOf(t, data["token"].(string))["roleId"].(float64); !ok || v != 3 {
		t.Errorf("claims roleId ไม่ใช่ 3")
	}
}

// TOKEN_EXPIRE ที่ใช้ไม่ได้ต้องถูกจับตั้งแต่ InitTokenExpire ตอนบูต
// ของเดิมตรวจข้างใน GetJWT แล้ว log.Fatalf = ตายทั้งโปรเซสตอนมีคน login
func TestInitTokenExpire_RejectsBadValues(t *testing.T) {
	for _, tc := range []struct{ name, value string }{
		{"ว่าง", ""},
		{"ไม่ใช่ตัวเลข", "abc"},
		{"ศูนย์", "0"},
		{"ติดลบ", "-1"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("TOKEN_EXPIRE", tc.value)
			if err := InitTokenExpire(); err == nil {
				t.Errorf("TOKEN_EXPIRE=%q ต้องคืน error แต่ผ่านไปได้", tc.value)
			}
		})
	}
}

// กันกรณีลืมเรียก InitTokenExpire ใน main — GetJWT ต้องคืน error ไม่ใช่เซ็น
// token ที่หมดอายุไปแล้วตั้งแต่วินาทีแรก
func TestGetJWT_MissingInitTokenExpire(t *testing.T) {
	t.Setenv("JWT_SECRET", "กุญแจทดสอบยาวพอสำหรับ HS256 อย่างน้อยสามสิบสองไบต์")
	if err := InitJwtSecret(); err != nil {
		t.Fatalf("InitJwtSecret: %v", err)
	}
	saved := tokenExpireDays
	tokenExpireDays = 0
	defer func() { tokenExpireDays = saved }()

	if _, err := GetJWT(models.AuthResultBranch{Data: &models.Users{ID: 1}}, "test"); err == nil {
		t.Error("ต้องคืน error เมื่อ tokenExpireDays ยังไม่ถูกตั้ง")
	}
}
