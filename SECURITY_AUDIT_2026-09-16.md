# pos-api-new — ผลตรวจสอบ 2026-09-16

ตรวจแบบอ่านอย่างเดียว ไม่แก้ไฟล์ ไม่รันเซิร์ฟเวอร์ ไม่เรียก endpoint ใด ๆ
ทุกข้อที่ระบุ "ยืนยันแล้ว" = อ่านโค้ดจริงที่บรรทัดนั้นด้วยตัวเอง ไม่ใช่การอนุมาน

**บริบท:** การตรวจรอบก่อน (ส.ค. 2026) ทำกับ `pos-api` ตัวเก่าที่ archive ไปแล้ว รอบนี้ตรวจ `pos-api-new` ซึ่งเป็นตัวที่ใช้งานจริง

---

## สรุปเทียบกับผลตรวจรอบเก่า

| ข้อค้นพบเดิม (จาก `pos-api` เก่า) | สถานะใน `pos-api-new` |
|---|---|
| Connection pool ไม่ได้ตั้งค่า | **แก้แล้วบางส่วน แต่กลายเป็นปัญหาใหม่** — ดูข้อ A1 |
| รหัสผ่าน plaintext | **ยังอยู่ และหนักกว่าเดิม** — ดูข้อ 1 |
| CORS เปิดกว้าง | ยังอยู่ (`cors.Default()`, `main.go:106`) แต่ความรุนแรงต่ำ — ดูท้ายเอกสาร |
| SQL injection | **ไม่พบ** — ตรวจทุก `fmt.Sprintf` ที่เข้า query แล้ว สะอาด |

---

## A. เสี่ยงต่อการ deploy โดยตรง (ต้องแก้ก่อนขึ้น)

### A1. Connection pool ใหม่อาจทำให้ production ล่มซ้ำรอยเมื่อวาน

`main.go:57-91`

```go
sqlDB1.SetMaxOpenConns(50)    // DB        (DATABASE_URL)
sqlDB2.SetMaxOpenConns(50)    // DB_POS    (DATABASE_POS_URL)
sqlDB3.SetMaxOpenConns(100)   // DB_ESTAMP (E_STAMP_DSN)
// DB_JREADER — ไม่ได้ตั้งค่าเลย → MaxOpenConns = ไม่จำกัด
```

เชื่อมต่อ 4 ฐานข้อมูล (`main.go:47-50`) แต่ตั้ง pool แค่ 3 ตัว **`DB_JREADER` ไม่ถูกตั้งค่า** จึงใช้ค่า default ของ Go คือ **MaxOpenConns ไม่จำกัด**

ทำไมเรื่องนี้สำคัญ: เหตุ POS ล่มเมื่อ 2026-09-15 เกิดจาก `posapidb` ที่ `max_connections = 100` ถูกใช้จนหมด ตอนนั้น pod `pos-api` (โค้ดเก่า) ใช้แค่ **2 connection** เพราะไม่เคยตั้ง pool เลย — จึงไม่ใช่ต้นเหตุ

แต่ `pos-api-new` ตั้ง 50 + 50 ต่อ pod และ HPA ให้ขยายได้ถึง 2 replica → **สูงสุด 200 connection** ต่อ 100 slot ที่มี บวก `DB_JREADER` ที่ไม่จำกัดอีก

**ต้องทำก่อน deploy:**
1. ตรวจว่า `DATABASE_URL` กับ `DATABASE_POS_URL` ของ production ชี้ไป `posapidb` ตัวเดียวกันหรือไม่ ถ้าใช่ = 100 connection ต่อ pod เต็มโควตาพอดี
2. ตั้ง pool ให้ `DB_JREADER` ด้วย
3. คำนวณย้อนกลับจากโควตาจริง: `MaxOpenConns × จำนวน pool ที่ชี้ DB เดียวกัน × max replica` ต้องน้อยกว่า `max_connections` โดยเหลือ slot ให้ superuser และเครื่องมือของคน

เกี่ยวข้อง: การตั้ง `idle_session_timeout` และแยก role อ่านอย่างเดียวสำหรับคน (ตามที่สรุปไว้ในเหตุเมื่อวาน) ยังจำเป็นอยู่ไม่ว่าจะแก้ข้อนี้หรือไม่

### A2. `.env` ถูก build ติดไปกับ Docker image

`Dockerfile:3` คือ `COPY . /app` และ **ไม่มีไฟล์ `.dockerignore`** (ตรวจแล้ว)

`.env` อยู่ใน `.gitignore` ถูกต้อง (git ไม่รั่ว) แต่ถูกคัดลอกเข้า image layer แล้วผลักขึ้น DO Container Registry ผ่าน CI ใครที่ pull image ได้จะได้ credential ทั้งหมดด้วย `docker run --entrypoint cat <image> /app/.env`

ใน `.env` มี: DSN ของ 4 ฐานข้อมูล (สองตัวเป็น superuser — `doadmin`, `root`), `JWT_SECRET`, `JOYLIDAY_AUTH`, `SMC_PWD`, `SERVICE_PASSWORD`, `NEWRELIC_LICENSE`

`tmp/main.exe` (62 MB) ก็ติดไปด้วย เพราะ `.gitignore` ไม่ครอบ `tmp/`

**แก้:** เพิ่ม `.dockerignore` ที่มี `.env`, `tmp/`, `.git` แล้ว**หมุน credential ทั้งหมดใหม่** — ต้องถือว่ารั่วไปแล้ว

### A3. JWT signing key อาจเป็นค่าว่าง — ⚠️ ต้องตรวจที่ pod จริง

`utils/Jwt.go:21`
```go
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))
```

เป็นตัวแปรระดับ package — Go สร้างค่านี้**ก่อน** `main()` ทำงาน ส่วน `godotenv.Load()` ถูกเรียกครั้งแรกใน `config.ConnectDatabase()` ที่ `main.go:47` คือหลังจากนั้นแล้ว

ถ้า production ได้ `JWT_SECRET` จาก `.env` เท่านั้น → `jwtSecret` เป็น `[]byte("")` และจะ**ไม่มีอาการผิดปกติใด ๆ** เพราะทั้งการเซ็นและการตรวจสอบใช้กุญแจว่างเหมือนกัน ระบบทำงานปกติทุกอย่าง แต่ใครก็ตามสร้าง token ปลอมเองได้ด้วยกุญแจว่าง

**แต่ถ้า** Kubernetes ฉีด `JWT_SECRET` เป็น env var จริง (ผ่าน Secret/ConfigMap ใน `infra.git`) ค่าจะถูกอ่านได้ตามปกติ และปัญหานี้ไม่เกิด — `godotenv.Load()` ไม่ทับ env var ที่มีอยู่แล้ว

**ยังสรุปไม่ได้จากโค้ดในเครื่อง** ต้องตรวจ 1 คำสั่ง:
```
kubectl -n apps exec deploy/pos-api -- printenv JWT_SECRET
```
ว่างเปล่า = ช่องโหว่ระดับวิกฤต ต้องหมุน secret และแก้ทันที

**ข้อดีที่ตรวจแล้วพบ:** `VerifyJwt` ตรวจ signing method อย่างถูกต้อง (`utils/Jwt.go:71`) — `alg=none` และ RS256→HMAC confusion **ทำไม่ได้** จุดนี้เขียนถูก

**แก้ไม่ว่าผลตรวจจะเป็นอย่างไร:** ย้ายการอ่าน secret เข้าไปในฟังก์ชัน และให้ `log.Fatal` ถ้าค่าว่างหรือสั้นกว่า 32 ไบต์

---

## B. ความปลอดภัยของบัญชีผู้ใช้ (ยืนยันแล้วทั้งหมด ไม่มีเงื่อนไข)

### 1. รหัสผ่านเก็บเป็น plaintext และ bcrypt ถูกคอมเมนต์ทิ้ง

`services/auth_service.go:24` และ `:53`
```go
if user.Password != password {
    return nil, errors.New("รหัสผ่านไม่ถูกต้อง")
}

// isValid := utils.CheckPasswordHash(password, user.Password)
// if !isValid {
//     return nil, errors.New("รหัสผ่านไม่ถูกต้อง")
// }
```

ฟังก์ชัน bcrypt มีอยู่และใช้งานได้ (`utils/Jwt.go:16-19`) แต่ **ไม่มีที่ไหนเรียกเลย** เขียนลง DB เป็น plaintext ตรง ๆ ที่ `services/user_service.go:72` และ `:117`

`UpdateUser` ยังเขียนทับรหัสผ่านโดยไม่เช็ค — ถ้า client ไม่ส่งฟิลด์ `password` มา จะกลายเป็นสตริงว่าง ล้างรหัสผ่านผู้ใช้ทิ้ง

### 2. API คืนรหัสผ่าน plaintext ของทุกคน

`models/user.go:52` — ไม่มี `json:"-"`
```go
Password string `gorm:"column:password" json:"password"`
```

มี 6 handler ที่คืน struct นี้ตรง ๆ โดยเฉพาะ **`GET /api/user/list`** ที่แบ่งหน้าได้และไม่กรองตามสิทธิ์ — พนักงานแคชเชียร์คนไหนก็ตามที่ login ได้ ยิง request เดียวก็ได้ username + รหัสผ่านของ admin ทั้งหมด

ฟิลด์นี้ปรากฏใน `models/user.go` บรรทัด 7, 18, 31, 52, 74

**แก้ได้ทันทีด้วยบรรทัดเดียว:** เปลี่ยนเป็น `json:"-"` — ควรทำก่อนอย่างอื่น เพราะเสี่ยงสูงและแก้ง่ายที่สุด

### 3. ไม่มีการตรวจสิทธิ์ฝั่งเซิร์ฟเวอร์เลย

`roleId` ถูกใส่ลง token ทุกใบ (`utils/Jwt.go:40`) แต่ค้นทั้ง codebase แล้ว **ไม่มีโค้ดบรรทัดใดอ่านค่านี้** ไม่มี `RequireRole` ไม่มี permission middleware ทุก route ผ่าน `AuthMiddleware()` ตัวเดียวที่เช็คแค่ "login แล้วหรือยัง"

ผลคือทุกคนที่ login ได้ ทำได้ทุกอย่าง:

| endpoint | ใครเรียกได้ |
|---|---|
| `POST /api/pos-void` | แคชเชียร์ทุกคน — และระบุ `voidUser` เป็นชื่อคนอื่นได้ (`pos_void_controller.go:156,172`) |
| `POST /api/refund-confirm` | แคชเชียร์ทุกคน |
| `POST /api/adjust-point` | แคชเชียร์ทุกคน — ไม่บันทึกด้วยซ้ำว่าใครทำ (ไม่เรียก `GetUserIdFromClaims` เลย) |
| `POST /api/user` | แคชเชียร์ทุกคน — ตั้ง `user_role_id` เป็น admin ให้ตัวเองได้ |
| `DELETE /api/card/:cardNo` | แคชเชียร์ทุกคน |

การขออนุมัติจากหัวหน้าก่อน void ที่มีในหน้าจอ POS เป็นการกั้นที่ **UI เท่านั้น** — ยิง API ตรงข้ามได้ทั้งหมด

### 4. Log รหัสผ่าน token และ JWT secret แล้วส่งออก New Relic

`controllers/auth_controller.go:29-30, 39`
```go
fmt.Println("jwtSecret", jwtSecret)
fmt.Println("JWT_SECRET", os.Getenv("JWT_SECRET"))
...
fmt.Println("username", req.Username, "password", req.Password)
```
`middlewares/auth.go:40` — `fmt.Println("Bearer token", token)` ทุก request

`main.go:51-55` เปิด `ConfigAppLogForwardingEnabled(true)` → stdout ทั้งหมดถูกส่งไปเก็บและทำดัชนีที่ New Relic (SaaS ภายนอก)

**ต้องทำ:** ลบบรรทัดเหล่านี้ และ **ล้าง log ของ New Relic ย้อนหลังตามช่วงที่เก็บไว้** เพราะมีรหัสผ่าน plaintext ของพนักงานทุกคนและ token ที่ยังใช้ได้อยู่

### 5. Basic auth ใช้ได้ทุก endpoint และทำให้ panic ได้

`middlewares/auth.go:56-89` รับ `Authorization: Basic` แทน JWT ได้ทุกเส้นทาง claim ที่สร้างขึ้นตั้ง `"exp": nil` และ**ไม่มี `roleId`** — session ไม่มีวันหมดอายุ

`middlewares/auth.go:65-66`
```go
userPass := strings.SplitN(string(decodedBytes), ":", 2)
auth, err := services.Authenticate(userPass[0], userPass[1])
```
ถ้าค่าที่ถอด base64 แล้วไม่มี `:` → slice ยาว 1 → `userPass[1]` panic (`gin.Recovery()` รับไว้เป็น 500) ส่ง `Authorization: Basic YQ==` โดยไม่ต้อง login ก็ทำให้ log ท่วมได้

**แก้:** `if len(userPass) != 2 { ... }`

---

## C. ความถูกต้องของเส้นทางเงิน (ยืนยันแล้วทั้งหมด)

**สาเหตุรากของเกือบทุกข้อในหมวดนี้มี 2 อย่าง:**
1. ทั้ง codebase มี **4 ไฟล์เท่านั้น** ที่ใช้ DB transaction และ **ไม่มี row locking เลยแม้แต่ที่เดียว** (`clause.Locking` / `FOR UPDATE` ค้นแล้วไม่เจอ)
2. การเช็คยอดคงเหลือทำใน Go จากค่าที่อ่านมาก่อนหน้า ไม่เคยเช็คใน SQL

### 1. หักยอดไม่สำเร็จแต่ระบบตอบว่าสำเร็จ → เล่นฟรีไม่จำกัด

`services/card_play_service.go:329-352`
```go
for _, WithDraw := range Withdraws {
    _, err := CreateCardWithdraw(...)
    if err == nil {
        _, err := UpdateCardDepositBalance(...)   // ← หักยอดอยู่ในนี้
        if err != nil { return nil, fmt.Errorf(...) }
    }
    // ← ไม่มี else
}
return Withdraws, nil   // ← คืน success เสมอ
```

การหักยอดอยู่ใน `if err == nil` และ**ไม่มี `else`** ถ้า `CreateCardWithdraw` ล้มเหลว `err` ถูกทิ้ง วนต่อ แล้ว return `nil` = สำเร็จ

**สถานการณ์จริง:** บัตรมี 500 coin → เรียก deduct 500 → INSERT `card_withdraw` ล้ม (FK, connection สะดุด) → ไม่หักยอด → API ตอบ 200 → เครื่องปลดล็อกให้เล่น → ยอดยังเป็น 500 **ทำซ้ำได้ไม่จำกัด และไม่มี row audit ให้ตรวจจับ**

### 2. Void ซ้ำได้ → ยอดติดลบ

`controllers/pos_void_controller.go:114-241`

goroutine ตัวแรกเช็ค `FindExistVoidByBillNo` ก่อนสร้าง record แต่ goroutine ตัวที่ 2, 3, 4 **ไม่ดูค่านั้นเลย** ทำงานทุกครั้งที่เรียก

**สถานการณ์จริง:** void บิล 500 coin สำเร็จ (ยอด 500→0) แต่ CRM sync timeout → `g.Wait()` คืน error → POS ขึ้น 500 "ล้มเหลว" → แคชเชียร์กด void ซ้ำ → goroutine 3 รัน `balance_coin = balance_coin - 500` บนบัตรที่เป็น 0 อยู่แล้ว → **ติดลบ 500** กดครั้งที่สาม → ติดลบ 1000 คะแนนสมาชิกก็ถูกหักซ้ำด้วย

`pos_void.bill_no` ไม่มี unique index จึงเป็น TOCTOU race ด้วย

### 3. Double-spend เพราะอ่าน-แล้ว-เขียนโดยไม่ล็อก

อ่าน+เช็ค: `services/card_play_service.go:778-797` → เขียน: `services/card_deposit_service.go:162`

```go
if balanceCoin >= int64(input.ECoin) {   // เช็คใน Go
...
// คนละที่ คนละ connection:
UPDATE card_deposit SET balance_coin = balance_coin - ? WHERE id = ?
```

UPDATE เป็น atomic ในตัวเอง แต่**เงื่อนไขการเช็คไม่ได้อยู่ใน WHERE** ไม่มี `FOR UPDATE` ไม่มี transaction ครอบ

**สถานการณ์จริง:** บัตรมี 100 → กดสองครั้งห่างกัน 50ms → A อ่าน 100 ผ่าน, B อ่าน 100 ผ่าน → A เขียน 0 → B เขียน **−100** ลูกค้าได้เล่น 2 ครั้งจากยอด 1 ครั้ง และ `SUM(balance_coin)` ที่ใช้ในรายงานทั้งหมดเพี้ยนตาม

**แก้แบบไม่ต้องล็อก:** ย้ายเงื่อนไขเข้า UPDATE → `WHERE id = ? AND balance_coin >= ?` แล้วเช็ค `RowsAffected == 1` วิธีนี้แก้ข้อ 3, 6, 8 พร้อมกัน

### 4. เติมเงินเข้าบัตรก่อนสร้างบิล + panic กลางทางจาก input ปกติ

`controllers/card_entity_controller.go:466-632`

ลำดับคือ: สร้าง `card_deposit` ใน goroutine ขนาน (เงินเข้าบัตรแล้ว) → `wg.Wait()` → **ค่อย**สร้าง `pos_transaction`

ระหว่างนั้นมี `*req.FreePoint` (บรรทัด 618) และ `uuid.MustParse(req.BillPaymenytId)` (บรรทัด 607) — `FreePoint` เป็น `*int` ที่**ไม่มี `validate:"required"`** และโค้ดเองพิสูจน์ว่าเป็น optional เพราะบรรทัด 635 เช็ค `req.FreePoint != nil` หลังจาก dereference ไปแล้ว

**สถานการณ์จริง:** POS ส่งเติมเงิน 500 โดยไม่ส่ง `free_point` → goroutine commit `card_deposit` **เงินเข้าบัตรแล้ว 500** → บรรทัด 618 nil-deref → panic → 500 → **ไม่มี `pos_transaction` ไม่มีเลขบิล ไม่มีบันทึกการชำระเงิน** ยอดนี้ void ไม่ได้ (`CreatePosVoid` ตอบ `billNo not found`) และไม่ปรากฏในรายงานใด ๆ แคชเชียร์เห็น error ก็เติมซ้ำ → อีก 500

### 5. เลขบิลซ้ำทุกวันเพราะลืมใส่วันที่

`services/pos_transaction_service.go:448-480`

นับจำนวนบิล scope เป็น "วันนี้" (`startOfDay`..`startOfNextDay`) แต่ format ใส่แค่ปีกับเดือน:
```go
BillNo := fmt.Sprintf("%s-%s-%02d%02d%04d", prefix, PosID, year%100, month, totalRecord)
```
ตัวแปร `day` ถูกคำนวณที่บรรทัด 452 แล้ว**ไม่ถูกใช้** (โค้ดเดิมที่คอมเมนต์ไว้บรรทัด 429-430 มี `DDStr` อยู่ — หายไปตอนเขียนใหม่) และ `bill_no` **ไม่มี unique index**

**สถานการณ์จริง:** 15 ก.ย. เครื่อง 01 ขาย 12 บิล → `FW-01-25090001`…`0012` วันที่ 16 ก.ย. ตัวนับรีเซ็ตเป็น 0 → บิลแรกได้ `FW-01-25090001` **ซ้ำกับของวันที่ 15** ทุกการค้นหาใช้ `First()` บน `bill_no` → **void บิลวันที่ 16 จะไปกลับรายการของลูกค้าวันที่ 15 แทน** เกิดขึ้นทุกวัน ไม่ต้องรอ concurrency

### 6. `redeem_price` ติดลบ = เสกเงิน

`controllers/jubu_jibi_controller.go:106-132` → `services/member_service.go:322-340`

validation มีแค่ `member_tel`, `location`, `sub_items` ต้องไม่ว่าง `RedeemPrice int` **ไม่มี validate tag** ไหลตรงเข้า:
```go
UPDATE member SET jubu_jibi = jubu_jibi - ? WHERE tel = ?
```
ไม่เช็คเครื่องหมาย ไม่เช็คยอดคงเหลือ ไม่มี floor (ต่างจาก `UpdatePoint` บรรทัด 232 ที่มี `CASE WHEN < 0 THEN 0`)

ส่ง `"redeem_price": -5000` → `jubu_jibi - (-5000)` = **เพิ่ม 5000** แล้ว sync ยอดที่เสกขึ้นมาไป CRM ต่อ (`member_service.go:150`)

### 7. Void เขียน 4 ที่พร้อมกันโดยไม่มี transaction และไม่ rollback

`controllers/pos_void_controller.go:110-246` ใช้ `errgroup.Group` (ไม่ใช่ `WithContext`) → **goroutine ตัวหนึ่งพังไม่ได้ยกเลิกตัวอื่น** ทุกตัววิ่งจนจบและ commit หมด

**สถานการณ์จริง:** goroutine 2 mark บิลเป็น VOID, goroutine 4 หักคะแนน + ลบประวัติ CRM, goroutine 3 พังกลางคัน (หัก deposit ได้ 1 จาก 3 ก้อน) → บัญชีบอกว่ากลับรายการครบ แต่ลูกค้ายังถือ coin เหลือ ~330 จาก 500

รูปแบบเดียวกันที่ `controllers/claim_prize_controller.go:50-93`

### 8. `adjust-point` เช็คยอดแบบ TOCTOU

`controllers/adjust_point_controller.go:35-58` — เป็นเส้นทางเดียวใน codebase ที่เช็คยอดก่อนหัก แต่เช็คใน memory แล้ว UPDATE ไม่มี `WHERE bonus >= ?` และไม่มี floor
สมาชิกมี 500 → admin สองคนกด −500 พร้อมกัน → ทั้งคู่ผ่าน → **−500** แล้ว sync ไป CRM

### 9. `if err != nil` กลับด้าน → panic แน่นอนบนเส้นทาง error

`services/card_deposit_service.go:166-178`
```go
deposit, err := FindCardDepositByIds(depositID)
if err != nil {                        // ← ควรเป็น == nil
    balance_coin := deposit[0].BalanceCoin   // ← deposit เป็น nil ตรงนี้เสมอ → panic
```
ผลสองทาง: (ก) ทางสำเร็จข้ามบล็อกนี้ → deposit ที่ยอดเหลือ 0 **ไม่เคยถูก mark ว่าใช้หมด** ยังโผล่เป็นแหล่งเงินที่ใช้ได้ต่อ (ข) ทาง error panic กลางการหักยอด

### 10. `discount_cash` ใช้ `float32`

`models/card_deposit.go:28-29, 52-53` — coin/bonus/point เป็น `int` ถูกต้องแล้ว แต่ discount cash เป็น `float32` (~7 หลักนัยสำคัญ) และถูกลบซ้ำ ๆ ด้วย `balance_discount_cash = balance_discount_cash - ?` ยอดที่ควรเป็น 0.00 พอดีจะไปจบที่ `-0.0000019` ทำให้การเทียบกับศูนย์ไม่ตรง

### 11. ไม่มี idempotency ที่ไหนเลย

ค้นทั้ง codebase: ไม่มี idempotency key ไม่มี request ID ไม่มี `uniqueIndex` tag ในโมเดลใดเลย

`TopupCardPOS` รับ `bill_no` มาจาก client ได้ (`models/card_entity.go:57`) และใช้ค่านั้นตรง ๆ **โดยไม่เช็คว่าซ้ำหรือไม่** → POS ส่งเติม 1000 สำเร็จแต่ response หาย → ส่งซ้ำ → **บัตรได้ 2000 จากการจ่ายเงินครั้งเดียว**

---

## D. ข้อที่ตรวจแล้วไม่พบปัญหา

- **SQL injection — ไม่พบ** ตรวจทุก `fmt.Sprintf` ที่เข้า query builder แล้ว ทั้ง `card_entity_service.go:1235` และ `meter_record_service.go:411` ต่อเฉพาะสตริงคงที่/placeholder ค่าจากผู้ใช้ผ่าน `args ...interface{}` ทั้งหมด ที่เหลือใช้ `?` placeholder
- **JWT algorithm confusion — ทำไม่ได้** `utils/Jwt.go:71` assert `*jwt.SigningMethodHMAC` ถูกต้อง `exp` ถูกตั้งและตรวจ
- **ไม่มี secret hardcode ใน .go เลย** ทุกค่าอ่านผ่าน `os.Getenv` — รูปแบบที่เคยพบใน frontend ไม่ได้เกิดซ้ำที่ backend
- **ไม่มีการจัดการเลขบัตรเครดิต/PAN/CVV** ในบริการนี้ การชำระเงินอ้างอิงเป็นรหัสประเภทเท่านั้น
- **`GetAllUsers`** (`services/user_service.go:18-22`) ทำถูกต้อง — `Select("id","name","s_name")` ลงใน `UserLite` ปัญหาอยู่ที่เส้นทางค้นหารายคนที่ไม่ได้ทำแบบเดียวกัน
- **`cors.Default()`** (`main.go:106`) เปิดทุก origin แต่ไม่ได้ตั้ง `AllowCredentials` และ `AllowHeaders` ไม่มี `Authorization` ด้วยซ้ำ — preflight ของเบราว์เซอร์จะไม่ผ่านสำหรับ request ที่แนบ token อยู่แล้ว ควรจำกัด origin แต่**ความรุนแรงต่ำกว่าที่รายงานรอบก่อนระบุไว้**
- **`middlewares/api_middleware.go`** สะอาด (log แค่ method+path) แต่ไม่เคยถูก register ใน `main.go` = dead code
- รายการ bypass auth มีแค่ 3 รายการ (`middlewares/auth.go:23`): `/api/swagger*`, `/api/auth/login`, `/logoes/` — ตรวจแล้วว่า path traversal ผ่าน prefix `/api/swagger` ไปถึง route ข้อมูลไม่ได้ (Gin ตั้ง `RedirectFixedPath: false`) **ไม่มี endpoint ที่แตะเงินหรือคืนข้อมูลเปิดทิ้งไว้แบบไม่ต้อง auth**
- Swagger UI เปิดสาธารณะโดยไม่มี env guard (`main.go:111`) — เป็นเพียงการเปิดเผยแผนผัง API ไม่ได้รั่วข้อมูล แต่ช่วยให้ผู้โจมตีทำงานง่ายขึ้นมาก

---

## ลำดับที่แนะนำให้แก้

**แก้ได้ทันที เสี่ยงต่ำ ผลตอบแทนสูง**
1. `models/user.go` — เปลี่ยน `json:"password"` เป็น `json:"-"` ทุกจุด (บรรทัด 7, 18, 31, 52, 74)
2. ลบ `fmt.Println` ที่ `controllers/auth_controller.go:29,30,39` และ `middlewares/auth.go:40,47`
3. เพิ่ม `.dockerignore` (`.env`, `tmp/`, `.git`)
4. `middlewares/auth.go:66` — ใส่ `if len(userPass) != 2`
5. `pos_transaction_service.go:472` — ใส่ `day` กลับเข้าไปในเลขบิล + เพิ่ม unique index บน `bill_no`
6. `services/card_play_service.go:340` — ใส่ `else { return nil, err }`
7. `services/card_deposit_service.go:167` — กลับเงื่อนไขเป็น `== nil`

**ต้องตรวจก่อนตัดสินใจ**
8. `kubectl exec deploy/pos-api -- printenv JWT_SECRET` → ถ้าว่าง = วิกฤต
9. ตรวจว่า DSN production ของ `DATABASE_URL` กับ `DATABASE_POS_URL` ชี้ฐานเดียวกันหรือไม่ แล้วปรับ pool

**ต้องวางแผนและทำ migration**
10. เปิด bcrypt กลับมา + migration hash รหัสผ่านเดิมทั้งหมด + หมุน credential ทุกตัวใน `.env` + ล้าง log New Relic ย้อนหลัง
11. ทำ role middleware แล้วบังคับใช้กับ void / refund / adjust-point / user admin
12. ย้ายเงื่อนไขเช็คยอดเข้า `WHERE` ของ UPDATE แล้วเช็ค `RowsAffected` — แก้ double-spend, redeem ติดลบ, adjust-point race พร้อมกัน
13. ครอบ transaction รอบ topup / void / claim-prize — ต้องรื้อ `errgroup`/`WaitGroup` ที่ fan-out อยู่ออกก่อน โค้ดมีตัวอย่างที่ถูกต้องแล้วที่ `services/clear_card_service.go:32` พร้อมคอมเมนต์กำกับว่า *"ทำ sequential ทีละอัน (ห้าม parallel บน tx เดียวกัน)"*
