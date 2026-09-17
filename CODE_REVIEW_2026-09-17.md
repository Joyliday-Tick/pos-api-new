# ผลรีวิวโค้ด 2026-09-17

รีวิวงานของวันที่ 17 ก.ย. ทั้งสอง repo (`pos-api-new` `b4c8991..fe8015a`,
`pos-fe-new` `d2516c4..b9e7292`) ด้วยผู้รีวิว 4 มุม: ไล่ทีละบรรทัด, ตรวจของที่ถูกลบ,
ไล่ผู้เรียกข้ามไฟล์, ความปลอดภัย และประสิทธิภาพ

ทุกข้อด้านล่างยืนยันด้วยการอ่านโค้ดจริงและ/หรือ query ข้อมูล UAT แบบอ่านอย่างเดียว

---

## แก้แล้ว

| # | เรื่อง | commit |
|---|---|---|
| 1 | `card_play_service.go:227` ใส่ `card_play.id` ลงช่อง `card_deposit id` → บัตร Package/Time-play เล่นฟรี | `e0917dc` |
| 2 | `releaseClaim` ถอนการจอง `pos_void` ทั้งที่คืนยอดไปแล้วบางส่วน → กดซ้ำแล้วหักซ้ำ | `20ff6ae` |
| 3 | `refund_confirm` ยืนยันซ้ำได้ + goroutine ขนานแบบไม่มีใครหยุดใคร | `20ff6ae` |

ข้อ 1 เป็น **regression ที่เกิดจาก `aff52a5` ของวันเดียวกัน** — guard `RowsAffected == 0`
ทำให้แถวที่ใส่ id ผิดตาราง (ซึ่งเดิมไม่มีพิษภัยเพราะ UPDATE ไม่โดนแถวไหนแล้วเงียบผ่านไป)
กลายเป็นตัวแรกของลูปที่คืน error เสมอ ของจริงจึงไม่เคยถูกหัก และผู้เรียกที่
`v2/card_entity_controller.go:220` รันแบบ fire-and-forget หลังตอบ success ไปแล้ว

หลักฐาน UAT: `card_deposit` ที่ id ตรงกับ `card_play.id` = **0 จาก 2678**,
`card_withdraw` ที่ `card_deposit_id` ไม่มีอยู่จริง = **221 แถว**,
`card_play_type` ที่ยังใช้งาน = **533 แถว**, foreign key บน `card_withdraw` = **ไม่มี**

---

## ยังไม่แก้ — เรียงตามความรุนแรง

### 1. Path traversal ที่ BFF → เลื่อนขั้นตัวเองเป็น Administrator ได้

`pos-fe-new/src/app/api/branch-group/[id]/route.js:39` และอีก **40 จุด** ยัดพารามิเตอร์
ลง URL ปลายทางโดยไม่ `encodeURIComponent`

ทดสอบจริงแล้ว: `new URL("http://backend:8080/api/branch-group/../user/42")`
ยุบเหลือ `/api/user/42`

แคชเชียร์ที่โดน 403 ที่ `/api/adjust-point` ส่ง
`PUT /api/branch-group/..%2Fuser%2F42` พร้อม body `{"user_role_id": 1}`
Next ถอดรหัส segment เป็น `../user/42` แล้ว `fetch` ยุบ path ไปโผล่ที่ `UpdateUser`
ซึ่งผ่าน `RequireRole(RoleAdministrator)` ได้ เพราะ BFF ยิงด้วย service account
`pos_frontend` ที่เป็น role 1 — **ลบล้าง RequireRole ทั้งชุดที่เพิ่งทำ**

route ที่รับค่าจาก query string (เช่น `card/check/package`) ง่ายกว่าอีก
เพราะส่ง `../` ตรง ๆ ได้โดยไม่ต้อง encode

**แก้:** ใส่ `encodeURIComponent` ทั้ง 40 จุด

### 2. route ที่ขยับยอดเงินแต่ไม่ได้ติด `RequireRole`

| route | ผล |
|---|---|
| `POST /api/member/:tel` (`UpdateSkill`) | `UPDATE member SET mskill1 = mskill1 + ?` ไม่มีเพดาน ไม่ตรวจเครื่องหมาย สร้างแต้มได้ไม่จำกัด ไม่มีแถว `adjust_point` จึงไม่โผล่ในรายงานที่หัวหน้าตรวจ |
| `POST /api/member/deduct-joylicoin/:tel` | ล้างยอด Joylicoin ของสมาชิกคนไหนก็ได้ |
| `POST /api/member/deduct-totalpoint/:tel` | เหมือนกัน และ `UpdateTotalPoint` ไม่เรียก `GetUserIdFromClaims` เลย = ไม่มีใครถูกบันทึกว่าเป็นคนทำ |
| `POST /api/card/clear-card` | ล้างมูลค่าทั้งใบ ขณะที่ `DELETE card/:cardNo` ซึ่งเบากว่ากลับถูก guard |
| `POST /api/claim-prize`, `POST /api/claim` | หักยอดบัตรผ่าน `CreateCardWithdraw`/`UpdateCardDepositBalance` แต่ไม่ถูก guard ทั้งที่ `refund-confirm` ข้าง ๆ ถูก guard |

### 3. `card_deposit_service.go:240` ปิด card play โดยดูแค่ coin

`deposit[0].BalanceCoin == 0` ไม่ดู `BalanceBonus` — โปรโมชันที่ให้ bonus อย่างเดียว
(`ECoin = 0`, `EBonus = 300`) จะถูกปิดสิทธิ์เล่นทั้งที่ยอด bonus ยังอยู่ครบ

บล็อกนี้ไม่เคยทำงานบนเส้นทางสำเร็จมาก่อน (เงื่อนไข `if err != nil` กลับด้าน แก้ใน `fbcd244`)
บั๊กจึงหลับอยู่ ตอนนี้ทำงานทุกครั้งที่หักยอด

### 4. `GenerateBillNo` ใช้คนละเขตเวลากับ `bill_date`

`pos_transaction_service.go:451` ใช้ `time.Now()` (เขตเวลา process) ขณะที่ `bill_date`
เขียนด้วย `utils.TimeNowAsia()` — ตรวจแล้ว container รัน **UTC** จริง (`TZ` ว่าง)

บิลที่ออกระหว่าง 00:00–07:00 ตามเวลาไทยจะมีวันที่ในเลขบิลเป็นของเมื่อวาน
ขณะที่ `bill_date` เป็นวันนี้ และตัวนับรีเซ็ตตอน 07:00 ไม่ใช่เที่ยงคืน
ขัดกับจุดประสงค์ของการแก้ C5

ที่อื่นก็ใช้ `time.Now()` แบบเดียวกันอีก ~10 จุด (`deposit_service.go:93`,
`redeem_service.go:86`, `card_play_service.go:409,508` …)

หมายเหตุพ่วง: `utils.TimeNowAsia()` เรียก `panic(err)` ถ้าโหลด `Asia/Bangkok` ไม่ได้
ตอนนี้ปลอดภัยเพราะ `Dockerfile` ใช้ `FROM golang:1.24` ซึ่งมี tzdata
**ถ้าย้ายไป base image เล็กลงตอน deploy จะ panic ทันที** ต้อง `import _ "time/tzdata"`
หรือติดตั้ง tzdata

### 5. หน้า adjust แถว E-Stamp ยังไม่ถูก

`pos-fe-new/src/app/(adjust)/adjust/page.jsx` — วันนี้แก้จาก `member.ecoin`
เป็น `member.estamp` ซึ่ง**ถูกแค่ครึ่งเดียว**

`adjust_point_controller.go:69-75` ตรวจเพดานการหัก `adj_estamp` กับ `findMember.Ecoin`
ส่วน `member_service.go` เอา `AdjEstamp` ไปบวกทั้ง `ecoin` และ `estamp`
สมาชิกที่ `estamp = 500` แต่ `ecoin = 100` จะเห็นหน้าจอ 500 ใส่ −500 แล้วได้
`400 Estamp ไม่เพียงพอ ในระบบมี 100` ซึ่งเป็นตัวเลขที่ไม่ปรากฏบนหน้าจอเลย

**ควรแสดงทั้งสองค่า**

### 6. `card_play_service.go` — สาขา `else` ที่เหลืออยู่

`claim_prize` และ `refund` แก้ให้ทำทีละคู่เรียงกันแล้ว แต่ `CreateClaimPrize`
กับตัวนับ JubuJibi ยัง commit ก่อนการหักยอดและไม่มีการชดเชย
ถ้าหักยอดล้มเหลวจะได้แถว claim ที่ไม่ได้หัก coin และกดซ้ำได้แถวใหม่เรื่อย ๆ

### 7. ความเร็ว

| จุด | ผล |
|---|---|
| `card_entity_controller.go:713` | เติมเงินให้สมาชิกที่มี free point เรียก CRM **5 ครั้งเรียงกัน** แต่ละครั้ง timeout 10 วิ = หน้าจอค้างได้ถึง **50 วินาที** แล้วตอบ 500 ทั้งที่เงินเข้าบัตรแล้ว `GetScoreType`/`GetBranchByCode` แทบไม่เปลี่ยน ควร cache |
| `main.go:108` | `MaxIdleConns = 2` ขณะที่ `MaxOpen = 10` — void ใช้ round-trip เรียงกัน 21–24 ครั้ง connection ส่วนเกินถูกปิดทิ้งแล้วต้อง handshake ใหม่ ≈ **+600ms ต่อ void** ควรตั้ง MaxIdle ใกล้เคียง MaxOpen |
| `card_deposit_service.go:237` | ทุกการหักยอดตามด้วย SELECT อีกหนึ่งครั้ง = 2–3 round-trip แทน 1 และยิงทุกครั้งที่แตะบัตรที่เครื่องเกม ใช้ `RETURNING` ได้ |
| `clearcard/page.jsx` | เคลียร์ทีละใบแบบ `await` ในลูป 10 ใบ ≈ 3 วินาที ทำขนานแบบจำกัดจำนวนได้ |

### 8. `Tables.jsx:81` — `editingQty` ค้างข้ามรายการขาย

React ไม่ยิง `onBlur` ตอน unmount และ `<Tables>` ถูกเรนเดอร์แบบไม่มีเงื่อนไข
(`topup/page.jsx:1443`) จึงไม่เคย unmount พิมพ์จำนวนแล้วลบแถวทิ้ง คีย์จะค้าง
แล้วไปทับจำนวนที่แสดงในรายการขายถัดไป — หน้าจอโชว์ 12 แต่คิดเงินจาก 1
(โค้ดที่เพิ่งแก้วันนี้)

### 9. `Jwt.go:84` — `roleId` เป็น `null` ได้

สาขา Bearer ใส่ `UserRoleId` ที่เป็น `*int` ลง claim ตรง ๆ ต่างจากสาขา Basic
ที่แก้ให้แปลง nil เป็น 0 แล้ว ผู้ใช้ที่ `user_role_id` เป็น NULL จะได้ 403
ที่อ่านแล้วเหมือนระบบพัง

**ต้องตรวจก่อน deploy:** service account ที่ `POS_LOCAL_BASE_URL` ใช้
(`loginLocalPos`) เป็นคนละ instance กับที่ตรวจไว้ ถ้า `user_role_id` ไม่ใช่ 1
ทั้ง `pos-void` `refund-confirm` `adjust-point` จะตอบ 403 ให้แคชเชียร์ทุกคนพร้อมกัน

### 10. PII ใน log

`[ORPHAN DEPOSIT]` (`card_entity_controller.go:663`) เขียนเลขบัตร ชื่อแคชเชียร์ สาขา
และ `[VOID]` (`pos_void_controller.go:299`) เขียนเบอร์โทรลูกค้า ลง stdout
ซึ่งไหลเข้า log stream ที่มีคนอ่านได้กว้างกว่าฐานข้อมูลและไม่มี retention policy
คำสั่งตามเก็บที่แนะนำไว้ใช้แค่ `bill_no` ก็พอ

---

## ตรวจแล้วสะอาด

- ลำดับ middleware ของ `RequireRole` ถูกต้อง และ fail closed ทุกทาง (`roleId = 0` ไม่ผ่าน)
- `Jwt.go` ตรึง signing method เป็น `*jwt.SigningMethodHMAC` — `alg=none` และ
  RS256→HMAC confusion ทำไม่ได้ และตรวจ `exp` จริง
- `/api/auth/login` ถูกข้ามด้วยการเทียบ path ตรง ๆ ก่อนเช็ค header
- ไม่มี route ซ้ำหรือ v2 ที่วิ่งไปถึง controller ที่ถูก guard
- `buildQuery` ใช้ `URLSearchParams` ไม่มีช่อง injection และแก้บั๊กคำว่า `"null"` ได้จริง
- `upstream.js` อ่าน body ครั้งเดียว ไม่ได้เพิ่ม latency และ service token ถูก cache ตาม `exp`
- API route ทั้ง 104 ตัวมี `apiGuard` ครบ ยกเว้น `auth/*` `device` `get-mac` `qz/*`
  ซึ่งเป็น endpoint ก่อน login โดยตั้งใจ
- `FindUserDbByUsername` กรอง `is_active AND NOT is_delete` — ผู้ใช้ role Disabled
  626 คน **login ไม่ได้เลยสักคน**
- CRM client ทุกตัวมี `Timeout: 10 * time.Second` (`crm_service.go`)
- `station/mac-address` ไม่ได้ถูกลบ (ลบแค่ helper ที่ตายแล้ว) และ `topupStore.js`
  เป็นไฟล์ว่างที่ไม่มีใคร import

---

## ที่ต้องทดสอบ (ยังไม่ได้ทำ)

ทั้งหมดนี้ยืนยันได้เฉพาะเส้นทางที่ถูกปฏิเสธ (409/400) เส้นทางที่ทำสำเร็จต้องเขียนข้อมูลจริง

1. **เล่นเกมด้วยบัตร Package หรือ Time-play** ← สำคัญสุด ตอนนี้เล่นฟรีอยู่ ดูว่ายอดลดจริง
2. ยกเลิกบิล — ควรเหมือนเดิม เปลี่ยนแค่เอาการถอนการจองออก
3. คืนเงิน — เปลี่ยนเยอะสุด และลองกดยืนยันซ้ำว่าได้ 409
4. แลกของรางวัลด้วยบัตร — ค้างจากรอบก่อน ยังไม่เคยทดสอบ
5. เติมเงินให้สมาชิกที่มี free point — เส้นทาง C4
