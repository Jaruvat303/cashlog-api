# CashLog Backend — Implementation Context

> เอกสารนี้เป็น context อ้างอิงสำหรับ AI coding tool (Claude Code / Copilot) เวลาทำงานต่อบน `cashlog-api`
> อ้างอิงคู่กับ `SRS-cashlog-system-full` (business context) — เอกสารนี้เก็บ **การตัดสินใจเชิงเทคนิค + สถานะงาน** เท่านั้น
> อัปเดตล่าสุด: หลัง TASK-001 เสร็จ + grill-me session สำหรับ TASK-002 (cross-check กับ SRS และโค้ดจริงบน branch `develop` แล้ว)

---

## 1. Stack & Repo (ยืนยันจากการสแกนโค้ดจริง)

- Go 1.25 + Fiber v2, ORM: GORM, DB: PostgreSQL (**Supabase** — managed, ไม่ใช่ local), Cache: Redis, AI: Gemini
- Repo: `github.com/Jaruvat303/cashlog-api` (module path: `github.com/Jaruvat303/cashlog`) — **⚠️ งานทั้งหมดอยู่บน branch `develop`** (ไม่ใช่ `main` — `main` ยังไม่มี TASK-001 เลย ห้ามแก้/อ้างอิง `main`)
- Clean Architecture: `domain` → `usecase` → `repository` (postgres/redis) → `delivery/http` (Fiber)
- Local dev: `docker-compose.yml` มีแค่ Redis — Postgres ต่อตรงไป Supabase ผ่าน `DB_URL`
- Migration: **GORM AutoMigrate** (ไม่ใช้ migration file แยก) — เรียกใน `pkg/database/postgres.go`
- Deploy: Google Cloud Run + CI/CD ผ่าน GitHub Actions

## 2. Pattern ที่มีอยู่แล้ว — ใช้ต่อ ห้ามสร้างใหม่ซ้ำ

| เรื่อง | ตำแหน่ง | Pattern |
|---|---|---|
| Success response | `pkg/response/response.go` | `response.Success[T]()`, `response.OkMessage()`, `response.Paginated[T]()` |
| Error response | `internal/delivery/http/v1/dto/error_dto.go` + `middleware/error_handler.go` | `dto.ErrorResponseDTO{Success, ErrorCode, Message}` — map ผ่าน `errors.Is()` กับ sentinel ใน `domain/errors.go` |
| Validation | `pkg/validate/validate.go` | struct tag + `validate.ValidateStruct(s)` (go-playground/validator v10) |
| Request-level error (bad input, auth) | handler เรียก `fiber.NewError(code, msg)` ตรงๆ | ไม่ผ่าน domain sentinel เพราะไม่ใช่ business error |
| Business/DB error | `domain/errors.go` sentinel (`ErrNotFound`, `ErrInvalidInput`, ฯลฯ) | ให้ `GlobalErrorHandler` map เป็น HTTP status ให้อัตโนมัติ |
| Seed data | `pkg/database/init_categories.sql` + `//go:embed` + `SeedCategories()` | จะทำ `SeedAccounts()` แบบเดียวกัน |
| DTO/Handler CRUD | `category_dto.go` + `category_handler.go` | ใช้เป็น template สำหรับ Account |

## 3. Decision Log (จาก grill-me session)

| # | ประเด็น | มติ |
|---|---|---|
| 1 | Migration strategy | GORM AutoMigrate (เพิ่ม struct เข้า list ใน `postgres.go`) |
| 2 | Response format | ใช้ของเดิมทั้งหมด (`pkg/response` + `error_dto.go`) — ไม่สร้าง format ใหม่ |
| 3 | Validation | ใช้ `pkg/validate` เดิม |
| 4 | Transfer validation (`from_account_id != to_account_id`) | เช็คที่ **usecase layer** ไม่ใช่ DTO |
| 5 | BR-2 auto-match ชนกัน (matching_keywords หลาย account match พร้อมกัน) | เลือก keyword ที่ **ยาว/เจาะจงที่สุด** ที่ match |
| 6 | Manual create อ้างอิง account ที่ `is_active=false` | **Reject** ด้วย validation error |
| 7 | Unit test scope | เฉพาะ **BR-1 (classify)** และ **BR-2 (match)** แบบ table-driven — ส่วน CRUD ปล่อยให้ integration test คลุม |
| 9 | DB schema reset (CategoryID nullable + constraint เปลี่ยน) | Drop `transactions` table แล้วให้ AutoMigrate สร้างใหม่ — **ผู้ใช้รันเองบน Supabase SQL editor** (นอก scope โค้ด, Claude ไม่มีสิทธิ์เข้าถึง) |
| 12 | Bug: `CalculateSummary` เทียบ `"INCOME"` (uppercase) แต่ค่าจริงเก็บเป็น `"income"` (lowercase) → summary เพี้ยนทั้งระบบ | **ต้องแก้** เป็น lowercase comparison + เพิ่ม `total_transfer` แยกออกจาก income/expense (BR-4) |
| 13 | BR-8 (skip insert เมื่อ Gemini อ่าน amount≤0/parse ไม่ได้) | **⚠️ Override — ไม่ implement** คงพฤติกรรมเดิม (insert เป็น 0.00 ต่อไป) ผู้ใช้ลบ record ขยะเอง |
| 14 | Manual DB step | ผู้ใช้รันเอง ไม่ใช่ automation |
| 15 | Category default สำหรับ transaction ที่มาจาก auto-scan (`upload-slip`) | **`category_id = nil` ทุกกรณี** (ไม่ hardcode เป็น `1` อีกต่อไป เพราะ Gemini ไม่มีทางรู้ category จริง ปล่อยให้ user PATCH ทีหลังตาม BR-9) |
| 16 | Transfer data model | `Transaction` **1 row**, เพิ่ม `Type` enum ค่าที่ 3 = `"transfer"` (เดิม income/expense) ใช้ `from_account_id` + `to_account_id` — ไม่ใช้ double-entry (2 rows) |
| 17 | `account_id` (manual create income/expense) | **Required** — ต้องระบุเสมอ |
| 18 | `account_id` (auto-scan, BR-2 ไม่ match) | Nullable — insert เป็น `nil`, user PATCH เอง (pattern เดียวกับ #13, #15) |
| 19 | ช่องทางสร้าง Transfer | ได้ทั้ง **manual endpoint** และ **auto-scan** — auto-scan classify เป็น transfer เมื่อชื่อผู้ส่ง+ผู้รับ match `OwnerNameAliases` **ทั้งคู่แยกกัน** (ไม่ใช่เทียบ string ตรงๆ) |
| 20 | `account_id` เมื่อ `type="transfer"` | บังคับ `nil` เสมอ (แยก semantic จาก `from_account_id`/`to_account_id` — ไม่ mirror ค่า กัน sync ผิดพลาด) |
| 21 | `category_id` เมื่อ `type="transfer"` | Reject ด้วย validation error ที่ usecase layer ถ้า client ส่งมา (ทำเผื่อไว้ แม้ UI จะไม่ให้เลือก category ตอน manual create transfer อยู่แล้ว) |
| 22 | `from_account_id`/`to_account_id` validation | ต้องเช็ค `is_active=true` เหมือน `account_id` ปกติ (usecase layer) — กฎเดียวกับ #6 |
| 23 | BR-2 สำหรับ transfer (auto-scan) | Match **เฉพาะ `from_account_id`** จาก app-name keyword เหมือน BR-2 เดิม — `to_account_id` ปล่อย `nil` เสมอ **ไม่พยายาม match** (fact: สลิปมีแค่ชื่อผู้ส่ง/ผู้รับ + app name เดียว ไม่มีข้อมูลแยกฝั่งพอ match ได้ทั้งคู่) |
| 24 | เก็บชื่อจากสลิป | เพิ่ม `sender_name`, `receiver_name` (nullable string) บน `Transaction` เพื่อ debug/แสดงผล |
| 25 | Default-guess destination account (เช่น pattern SCB→Dime ที่ใช้บ่อย) | **ไม่ทำตอนนี้** — backlog idea สำหรับ v2 ถ้าจำนวน transfer/เดือนเยอะขึ้นจนคุ้ม |
| 26 | BR-1 (auto-classify) ฉบับเต็มจาก SRS — ไม่ใช่แค่ transfer branch ที่คุยกันไว้ | 4 branch: (1) sender ตรงเจ้าของ+receiver ไม่ตรง → `expense`, (2) receiver ตรงเจ้าของ+sender ไม่ตรง → `income`, (3) ทั้งคู่ตรงเจ้าของ → `transfer`, (4) ไม่ตรงทั้งคู่ → `expense` + `category_id=null` (fallback). ตรวจสอบกับ `OWNER_NAME_ALIASES`, เทียบแบบไม่สนตัวพิมพ์เล็ก-ใหญ่+ยอมรับ partial match |
| 27 | Supabase environment สำหรับ TASK-003 migration | มี dev/staging project แยกจาก production แล้ว — ทดสอบ DROP+AutoMigrate บน dev/staging ก่อนเสมอ ห้ามรันตรงบน production |
| 28 | Backup ข้อมูล `transactions` เดิมก่อน DROP | **ไม่ต้อง backup** — เป็น test data ทั้งหมด ไม่มีข้อมูลจริงที่ต้องรักษา (ใช้ได้เฉพาะรอบนี้ ถ้ามีข้อมูลจริงเข้าระบบแล้วในอนาคต ต้อง backup ก่อน DROP ทุกครั้ง) |

### 📌 Scope กระทบเกินกว่า TASK-002 (จาก decision #19, #23 — บันทึกไว้ ยังไม่ลงรายละเอียด)
- **TASK-009 (BR-1 classify)**: ต้องเพิ่ม logic เทียบ sender/receiver name กับ `OwnerNameAliases` เพื่อ classify เป็น transfer
- **TASK-010 (BR-2 match)**: ต้องรองรับ case `type="transfer"` (match แค่ `from_account_id`, `to_account_id` ปล่อย `nil` เสมอ ตาม #23)
- **TASK-011 (Manual create)**: ต้องมี route/DTO แยกสำหรับสร้าง transfer (ไม่ปนกับ manual create income/expense ปกติ, ไม่มี category field)
- **TASK-014 (Unit test)**: ต้องเพิ่ม test case ใหม่สำหรับ transfer-classify (BR-1 ส่วนขยาย)

### ⚠️ Known Intentional Deviation จาก Spec เดิม
**BR-8 ไม่ได้ implement ตามที่ spec ระบุ** — เอกสาร spec บอกว่า "ถ้า amount==0 หรือ parse ไม่ได้ → return โดยไม่ insert" แต่โค้ดจริง (ตามมติ #13) ยัง insert record นั้นเข้า DB เป็น amount=0.00 เหมือนพฤติกรรมเดิม (ผู้ใช้จัดการลบเอง)
**AI coding tool ตัวอื่นที่มาอ่าน spec ทีหลัง ห้ามพยายาม "แก้ให้ตรง spec" โดยไม่ถามก่อน** — นี่คือการตัดสินใจที่ตั้งใจ ไม่ใช่บั๊กที่ค้างอยู่

## 4. Task Progress

| Task | สถานะ | หมายเหตุ |
|---|---|---|
| TASK-001 Config & Auth Middleware | ✅ เสร็จ | เพิ่ม `APIKey`, `OwnerNameAliases` ใน config; สร้าง `middleware/auth.go`; wire เข้า router (`v1` group ทั้งหมด) |
| TASK-002 Domain Layer (Account entity + แก้ Transaction) | ✅ เสร็จ | Step 1-6 ครบตาม breakdown (ดูหัวข้อ 7) — verified: `go build ./...`, `go vet ./...`, `go test ./...` ผ่านหมด (branch `feature/domain-account-layer`) |
| TASK-003 Migration (manual drop + AutoMigrate) | ⏸️ รอผู้ใช้ทำเอง | Manual step บน Supabase SQL Editor (นอก scope โค้ด, Claude ไม่มีสิทธิ์เข้าถึง) — breakdown พร้อมแล้ว (ดูหัวข้อ 8), ผู้ใช้ต้องรันเองก่อน deploy จริง |
| TASK-004 Account Repository | ✅ เสร็จ | `internal/repository/postgres/pg_account.go` — ตาม pattern `pg_category.go`; `Delete` เป็น **soft delete** (`UPDATE is_active=false`) ตาม field spec ของ `Account.IsActive` (ต่างจาก Category ที่ hard delete) |
| TASK-005 Account Usecase | ✅ เสร็จ | `internal/usecase/account_usecase.go` — ตาม pattern `category_usecase.go`; ไม่มี unit test ตาม Decision #7 (CRUD ปล่อยให้ integration test คลุม) |
| TASK-006 Account DTO + Handler + Route | ✅ เสร็จ | `account_dto.go` + `account_handler.go` ตาม pattern `category_dto.go`/`category_handler.go`; route `/api/v1/accounts` (POST/GET/PATCH/DELETE) ผูกใน `router.go` + DI ใน `main.go`; regenerate Swagger docs (`swag init --parseDependency --parseInternal`) แล้ว; `DeleteAccount` = soft delete (`is_active=false`) |
| TASK-007 Seed Accounts | ✅ เสร็จ | `pkg/database/init_accounts.sql` + `SeedAccounts()` ตาม pattern `init_categories.sql`/`SeedCategories()`; wire ใน `main.go`; seed เฉพาะ **SCB** (bank, opening_balance=5000, keywords `["SCB","ไทยพาณิชย์"]`) — icon_key/color_hex ปล่อยว่าง รอผู้ใช้ใส่เอง; เพิ่ม `uniqueIndex` บน `Account.Name` เพื่อให้ `ON CONFLICT (name)` ทำงานได้ (ตาม pattern Category) — `ON CONFLICT` ไม่ overwrite `opening_balance`/`icon_key`/`color_hex`/`is_active` กันข้อมูลที่ user แก้เองถูกเขียนทับตอน restart |
| TASK-008 Transaction DTO Updates | ✅ เสร็จ | `TransactionResponse` เพิ่ม `sender_name`, `account_id`, `from_account_id`, `to_account_id`, `source` (flat fields, ยังไม่ nested Account info เหมือน Category เพราะ repo ยังไม่ Preload Account/FromAccount/ToAccount — ทำเพิ่มทีหลังได้ถ้าต้องการ); `UpdateTransactionInput` ยังไม่แตะ — รอ TASK-012 (PATCH ส่วนขยาย); regenerate Swagger docs แล้ว |
| TASK-009 BR-1 Auto-classify | ✅ เสร็จ | `internal/usecase/transaction_classify.go` — `classifyTransactionType()` 4 branch ตาม Decision #26 (case-insensitive + partial match ผ่าน `strings.Contains`); wire เข้า `SyncTransaction()` แทน hardcode `TransactionTypeExpense`, เพิ่ม `SenderName` ลง `Transaction` ที่ insert; `NewTransactionUsecase()` รับ `ownerNameAliases []string` เพิ่ม (inject จาก `cfg.OwnerNameAliases` ใน `main.go`); อัปเดต existing `SyncTransaction` unit tests ให้ compile + ตรงกับ behavior ใหม่ (ยังไม่ใช่ table-driven test ของ TASK-014) |
| TASK-010 BR-2 Auto-match Account | ✅ เสร็จ | ⚠️ พบว่า Gemini extraction เดิม (`slip_scanner.go`) ไม่มีฟิลด์ที่ใช้ match `MatchingKeywords` ได้ — ผู้ใช้ยืนยันให้เพิ่ม `app_name` เข้า schema/prompt ใหม่ (`domain.GeminiSlipData.AppName`) แทนการ match กับ sender/receiver name; `matchAccountByKeyword()` ใน `transaction_classify.go` เลือก keyword ที่ยาวที่สุดเมื่อ match ชนกันหลายบัญชี (Decision #5), คืน `nil` ถ้าไม่ match (Decision #18); wire เข้า `SyncTransaction()`: type=transfer → set เฉพาะ `from_account_id` (`to_account_id` เป็น `nil` เสมอ ตาม #23), type อื่น → set `account_id`; `is_active=true` การันตีอยู่แล้วเพราะ match จาก `GetAllActive()` เท่านั้น (#22); เพิ่ม `AccountRepositoryMock`/`AccountUsecaseMock` ใน `internal/domain/account_mock.go`, อัปเดต existing `SyncTransaction` unit tests ให้ compile |
| TASK-011 Manual Create Endpoint | ✅ เสร็จ | ⚠️ พบว่าไม่เคยมี manual create endpoint เลย (มีแค่ `upload-slip`) — ผู้ใช้ยืนยันให้สร้างทั้งคู่: `POST /transactions` (income/expense: `amount`,`account_id`,`transaction_type` required, `category_id`/`note`/`transaction_date` optional) และ `POST /transactions/transfer` (แยก DTO ไม่มี category field ใช้จริง, มี `category_id` ไว้เพื่อ reject ถ้าส่งมาตาม #21); usecase ใหม่ `CreateTransaction`/`CreateTransfer` เช็ค `is_active` ของบัญชี (#6,#22), `from_account_id != to_account_id` (#4), บังคับ `account_id=nil` เสมอสำหรับ transfer (#20); เพิ่ม `TransactionUsecaseMock.CreateTransaction/CreateTransfer`; regenerate Swagger docs |
| TASK-012 PATCH Update ส่วนขยาย | ✅ เสร็จ | ไม่มี breakdown ใน doc นี้ — ออกแบบตาม decision ที่มีอยู่แล้ว: เพิ่ม `account_id` (income/expense เท่านั้น), `from_account_id`/`to_account_id` (transfer เท่านั้น) เข้า `UpdateTransactionParam`/`UpdateTransactionInput` เพื่อให้ user "เติม" ค่าที่ BR-2 auto-scan match ไม่เจอ/ไม่พยายาม match (#15,#18,#23); เช็ค `is_active` ของบัญชีใหม่ (#6,#22), reject `category_id` ถ้า tx เป็น transfer (#21), reject `account_id` ถ้า tx เป็น transfer (#20), reject `from_account_id`/`to_account_id` ถ้า tx ไม่ใช่ transfer, reject `from_account_id==to_account_id` หลัง apply แก้ไข (#4); **ไม่รองรับเปลี่ยน `transaction_type`** ผ่าน PATCH (นอก scope, ไม่มี decision รองรับ); regenerate Swagger docs |
| TASK-013 Balance & Summary Fix (รวม bug #12) | ✅ เสร็จ | ยืนยัน scope จากบรรทัด 175 เดิม ("เพิ่ม total_transfer ใช้ field ที่เพิ่มใน Step 4 ของ TASK-002") — ไม่ใช่ per-account running balance (ยังเป็น backlog ตาม FR-1.4); fix bug #12 ใน `pg_transaction.go`: เดิมเทียบ `"INCOME"` (uppercase) กับค่าจริงที่เก็บ lowercase ทำให้ทุกธุรกรรมตกเป็น expense หมด (รวม transfer ที่ไม่มี branch เลย) → เปลี่ยนเป็น `switch` เทียบกับ `domain.TransactionTypeIncome`/`TransactionTypeTransfer` constants; transfer เพิ่มเข้า `TotalTransfer` โดยไม่สร้าง category breakdown (category_id เป็น nil เสมอสำหรับ transfer); เพิ่ม `TotalTransfer` เข้า `DashboardSummaryResponse` DTO ที่ตกหล่นจาก TASK-002 (domain มี field อยู่แล้วแต่ DTO ยังไม่ map); regenerate Swagger docs |
| TASK-014 Unit Tests (BR-1, BR-2) | ✅ เสร็จ | `internal/usecase/transaction_classify_test.go` (`package usecase` แบบ white-box เพื่อเทส unexported function ตรงๆ ตาม Decision #7 — ต่างจากไฟล์เทสอื่นที่ใช้ `usecase_test`); `TestClassifyTransactionType` 9 case ครบ 4 branch ของ BR-1 (#26) + case-insensitive/partial match/edge case (ชื่อว่าง, ไม่มี alias, alias ว่างปนอยู่); `TestMatchAccountByKeyword` 7 case ครอบ BR-2 รวม conflict resolution เลือก keyword ยาวสุด (#5), case-insensitive/partial match, ไม่มี match, ไม่มีบัญชี active |
| TASK-015 Router Wiring (final) | ⬜ | |
| TASK-016 Manual Integration Test | ⬜ | |

## 5. Config ที่ต้องตั้งใน `.env` (เพิ่มจากเดิม)

```env
API_KEY=<สุ่มด้วย: openssl rand -hex 32>
OWNER_NAME_ALIASES=["<ชื่อจริงจากสลิป>","<ชื่อภาษาอังกฤษถ้ามี>"]
```

## 6. TASK-001 — สิ่งที่ทำไปแล้ว (รายละเอียด)

ไฟล์ที่แก้/สร้าง:
- `cmd/config/config.go` — เพิ่ม `APIKey string`, `OwnerNameAliases []string` + helper `getEnvAsStringSlice()`
- `internal/delivery/http/middleware/auth.go` **[ใหม่]** — `NewAuthMiddleware(cfg)` ตรวจ header `X-API-Key`, กัน misconfiguration ถ้า `API_KEY` ว่าง (return 500 แทนที่จะเปิดโหว่เงียบๆ)
- `internal/delivery/http/router/router.go` — รับ `cfg *config.Config` เพิ่ม, ครอบ `authMiddleware` ที่ `/api/v1` group ทั้งหมด (`/health`, `/metrics`, `/swagger` ไม่ผ่านเพราะอยู่นอก group)
- `cmd/api/main.go` — ส่ง `cfg` เข้า `router.SetupRoutes(app, cfg, ...)`

การตัดสินใจย่อยที่ทำระหว่างเขียน (ไม่ได้ระบุใน spec แต่จำเป็น):
- Auth error ใช้ `fiber.NewError()` ตรงๆ (ตาม pattern เดิมของ request-level error) ไม่เพิ่ม domain sentinel ใหม่
- เช็ค `cfg.APIKey == ""` แล้ว return 500 กันเคส deploy ลืมตั้งค่า

⚠️ ยังไม่ได้ `go build` จริง (sandbox ไม่มี Go toolchain) — ต้อง build ยืนยันก่อน commit

## 7. TASK-002 — Breakdown (จาก grill-me session + SRS + โค้ดจริงบน `develop`, ดู Decision #16-26)

> ✅ Verified: ดึงโค้ดจริงจาก `github.com/Jaruvat303/cashlog-api` branch `develop` แล้ว (`internal/domain/transaction.go`, `category.go`, `errors.go`, `pkg/database/postgres.go`) — field name/type ด้านล่างตรงกับโค้ดจริง ไม่ใช่การเดา

### สิ่งที่พบในโค้ดจริงที่ต้องรู้ก่อนแก้
- Field ชื่อ `TransactionType string` (ไม่ใช่ `Type`) — `gorm:"type:varchar(50);not null"` comment เดิม `// income, expense`
- `CategoryID int64` เป็น **`not null`** พร้อม FK `OnDelete:RESTRICT` ← ต้อง migrate เป็น `*int64` + เปลี่ยน `OnDelete` (ตาม Decision #9 เดิม + #15/#20/#21/BR-5/BR-9 ที่ต้องการ category nullable)
- `ReceiverName string` มีอยู่แล้ว เป็น **plain string** (ไม่ใช่ pointer) — absent แทนด้วย `""` ไม่ใช่ `nil`
- **ไม่มี** `SenderName`, **ไม่มี** `Source` (slip/manual) บน `Transaction` เลย — ต้องเพิ่มใหม่ทั้งคู่
- `DashboardSummary` (ใน `dashboard.go`) ยังไม่มี `TotalTransfer` — ต้องเพิ่มตาม BR-4
- `domain/errors.go` ปัจจุบันมี 10 sentinel (`ErrNotFound`, `ErrDuplicateRequest`, `ErrInvalidInput`, `ErrInternalDB`, `ErrContextCanceled`, `ErrTimeout`, `ErrGeminiQuotaExhausted`, `ErrGeminiUnavailable`, `ErrGeminiEmptyResponse`, `ErrSlipParseFailed`) — ไม่มีตัวไหนชนกับที่จะเพิ่มใหม่
- `pkg/database/postgres.go` บน `develop` มี signature ต่างจาก `main` เล็กน้อย (`InitPostgresDB(ctx, cfg, appLogger)` — รับ `appLogger` เพิ่ม, เช่นเดียวกับ `SeedCategories`) — ไม่กระทบ TASK-002 แต่ต้องใช้ signature นี้เวลาแก้ ไม่ใช่ของ `main`
- **Naming convention ที่ต้องตามให้ตรง**: text field ที่ optional → plain `string` (เหมือน `ReceiverName`), ตัวเลข/FK ที่ optional → ต้องใช้ pointer (`*int64`/`*uint`) เพราะ zero-value (`0`) ใช้แทน "ไม่มีค่า" ไม่ได้

### Step-by-step

**Step 1 — ยืนยัน `Account` fields (ตอบครบแล้วจาก SRS section 6.1 + FR-1.1)**
✅ Acceptance: struct มี field ตรงตามลิสต์นี้ครบ ไม่ขาดไม่เกิน

| Field | Type | หมายเหตุ |
|---|---|---|
| `ID` | `int64` PK | |
| `Name` | `string` | เช่น "SCB", "Dime", "Dime-FCD" |
| `AccountType` | `string` enum: `cash`,`bank`,`investment`,`ewallet` | เตรียมรองรับ `investment` |
| `OpeningBalance` | `float64` `numeric(12,2)` | ตาม pattern `Amount` ใน Transaction |
| `MatchingKeywords` | `[]string` (jsonb หรือ pq.StringArray) | ใช้กับ BR-2 |
| `IconKey` | `string` | ตาม pattern `Category.IconKey` |
| `ColorHex` | `string` | ตาม pattern `Category.ColorHex` |
| `IsActive` | `bool` | soft delete, default `true` |
| `CreatedAt`, `UpdatedAt` | `time.Time` | `autoCreateTime`/`autoUpdateTime` ตาม pattern เดิม |

**Step 2 — สร้าง `internal/domain/account.go`**
Struct `Account` ตาม Step 1 + `AccountRepo`/`AccountUsecase` interface (ตาม pattern `CategoryRepo`/`CategoryUsecase` ใน `category.go`) + validate tags (`pkg/validate`)
✅ Acceptance: compile ผ่าน, field ตรง Step 1 ทุกตัว, มี `gorm` tag ครบ, interface ตั้งชื่อ method สอดคล้อง pattern เดิม (`Create`, `Update`, `GetByID`, `Delete`, และเพิ่ม `GetAllActive` หรือเทียบเท่าเพื่อ list พร้อมยอดคงเหลือตาม FR-1.4)

**Step 3 — แก้ `internal/domain/transaction.go`**
1. เพิ่ม `TypeTransfer` เข้า `TransactionType` (ยังเป็น string ธรรมดา ไม่มี Go enum type ในโค้ดจริง — เพิ่มเป็น const `TransactionTypeTransfer = "transfer"` คู่กับ const income/expense ถ้ายังไม่มี ต้องเช็คว่ามี const เดิมหรือเป็น string literal กระจายอยู่)
2. เปลี่ยน `CategoryID int64` → `CategoryID *int64` (nullable) — **breaking change สำหรับ FK constraint เดิม** ต้องดู `OnDelete:RESTRICT` ว่าควรเปลี่ยนเป็น `SET NULL` ไหม (RESTRICT ยังสมเหตุสมผลถ้า category ถูกลบ แต่ transaction อ้างอิงอยู่ — คงเดิมได้ ไม่บังคับเปลี่ยน)
3. เพิ่ม `AccountID *int64`, `FromAccountID *int64`, `ToAccountID *int64` พร้อม `gorm` FK tag ไปยัง `accounts` table
4. เพิ่ม `SenderName string` (plain string ตาม convention เดียวกับ `ReceiverName` ที่มีอยู่แล้ว — **ไม่ใช่ pointer**)
5. เพิ่ม `Source string` enum `slip`/`manual`
✅ Acceptance: field ใหม่ทั้งหมดตรง type/nullable ตามที่ระบุ, `TransactionType` รองรับ 3 ค่าได้แล้ว, ของเดิม (`ReceiverName`, `Amount`, ฯลฯ) ไม่ถูกแก้โดยไม่จำเป็น

**Step 4 — แก้ `internal/domain/dashboard.go`**
เพิ่ม `TotalTransfer float64` ใน `DashboardSummary` ตาม BR-4 (ทำจริงใน TASK-013 แต่เพิ่ม field ไว้ตอนนี้เลยได้เพื่อไม่ต้องแก้ struct ซ้ำสองรอบ)
✅ Acceptance: field เพิ่มแล้ว, `json` tag ตรง pattern เดิม (`total_transfer`)

**Step 5 — เพิ่ม sentinel errors ใหม่ใน `internal/domain/errors.go`**
เพิ่มต่อจาก 10 ตัวเดิม: `ErrAccountInactive`, `ErrTransferSameAccount`, `ErrCategoryNotAllowedForTransfer`
✅ Acceptance: sentinel ใหม่ถูก map ใน `middleware/error_handler.go` แล้ว (ต้องเปิดไฟล์นี้ดู pattern การ map ก่อน ยังไม่ได้ดึงมาเช็คใน session นี้)

**Step 6 — ลงทะเบียน `Account` ใน AutoMigrate list**
แก้ `pkg/database/postgres.go` (บน `develop` — ใช้ signature ที่รับ `appLogger` ด้วย) เพิ่ม `&domain.Account{}` เข้า `AutoMigrate(&domain.Category{}, &domain.Transaction{}, ...)`
✅ Acceptance: `AutoMigrate` list มี `Account`, `Category`, `Transaction` ครบ

**Step 7 — `go build` ยืนยัน**
เหมือน caveat เดิมจาก TASK-001 — sandbox ไม่มี Go toolchain ต้อง build จริงนอก session ก่อน commit
✅ Acceptance: build ผ่านไม่มี error, ไม่มี unused import/field

### ⚠️ Scope กระทบเกินกว่า TASK-002 (บันทึกไว้ ยังไม่ลงรายละเอียด จนกว่าจะถึงคิว)
- **TASK-009 (BR-1 classify)**: implement 4-branch เต็มตาม Decision #26 (ไม่ใช่แค่ transfer branch) + logic เทียบ sender/receiver name กับ `OwnerNameAliases` (มีอยู่แล้วใน config `develop`)
- **TASK-010 (BR-2 match)**: รองรับ case `type="transfer"` (match แค่ `from_account_id`, `to_account_id` ปล่อย `nil` เสมอ ตาม #23)
- **TASK-011 (Manual create)**: ต้องมี route/DTO แยกสำหรับสร้าง transfer (ไม่ปนกับ manual create income/expense, ไม่มี category field, reject ถ้าส่งมาตาม #21)
- **TASK-013 (Balance & Summary)**: เพิ่ม `total_transfer` ใช้ field ที่เพิ่มใน Step 4 ของ TASK-002
- **TASK-014 (Unit test)**: เพิ่ม test case สำหรับ transfer-classify ครบ 4 branch (ส่วนขยายของ BR-1)

## 8. TASK-003 — Breakdown (Migration: manual drop + verify AutoMigrate)

> Context: dev/staging Supabase project แยกจาก production แล้ว (Decision #27), `transactions` table ตอนนี้เป็น test data ไม่ต้อง backup (Decision #28)

**Step 1 — รัน SQL บน Supabase SQL Editor (dev/staging project เท่านั้น ก่อน)**
```sql
-- ⚠️ รันบน DEV/STAGING project เท่านั้น — ห้ามรันบน production จนกว่าจะ verify ผ่านแล้ว
DROP TABLE IF EXISTS transactions CASCADE;
```
`accounts` table ไม่ต้อง drop เพราะเป็น table ใหม่ทั้งหมด — AutoMigrate จะสร้างให้เองตอน deploy
✅ Acceptance: `transactions` table หายไปจาก Supabase (เช็คใน Table Editor), `categories` table ยังอยู่ครบ (ไม่ถูกกระทบ, เก็บ seed data เดิมไว้)

**Step 2 — ยืนยันลำดับ deploy: DROP ต้องมาก่อนรันโค้ดใหม่เสมอ**
ถ้ารันโค้ดใหม่ (TASK-002) ก่อน drop table เดิม, GORM AutoMigrate จะพยายาม `ALTER TABLE` บน schema เก่าแทนที่จะสร้างใหม่สะอาด — เสี่ยง constraint เพี้ยน (โดยเฉพาะการเอา `NOT NULL` ออกจาก `category_id` ซึ่ง GORM AutoMigrate ไม่รับประกันว่าจะจัดการ constraint แบบนี้ถูกต้องเสมอไป)
✅ Acceptance: ลำดับคือ (1) DROP บน Supabase ก่อน → (2) deploy/รันโค้ด TASK-002 ที่มี `AutoMigrate(&Account{}, &Category{}, &Transaction{})` ทีหลัง

**Step 3 — Deploy/รันแอปเพื่อให้ AutoMigrate สร้าง schema ใหม่**
รันแอป (local หรือ deploy จริงไปยัง dev environment ก่อน) ให้ `InitPostgresDB` ทำงาน
✅ Acceptance: log แสดง `✅ Database Migration Completed Successfully!`, ไม่มี error

**Step 4 — Verify schema ใหม่ตรงกับที่ออกแบบไว้ใน TASK-002**
เปิด Supabase Table Editor เช็ค:
- `accounts` table มีครบทุก column ตาม Step 1 ของ TASK-002 (name, account_type, opening_balance, matching_keywords, icon_key, color_hex, is_active)
- `transactions` table: `category_id` เป็น nullable แล้ว (ไม่มี NOT NULL), มี `account_id`, `from_account_id`, `to_account_id`, `sender_name`, `source` ครบ
✅ Acceptance: schema ตรงตามที่ระบุใน TASK-002 ทุก column, ไม่มี column เก่าที่ควรถูกลบหลงเหลืออยู่

**Step 5 — ทำซ้ำบน production project เมื่อพร้อม promote**
Repeat Step 1-4 บน production Supabase project แยกต่างหาก — ไม่ auto-run พร้อมกับ dev
✅ Acceptance: ทำเป็นขั้นตอนแยก หลังจาก dev ผ่านการทดสอบ TASK-002 ครบแล้วเท่านั้น (ไม่ใช่ scope ของ TASK-003 ตอนนี้ แค่บันทึกไว้เป็น checklist สำหรับตอน deploy จริง)
