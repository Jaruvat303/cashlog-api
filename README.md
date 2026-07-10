# CashLog API

## ภาพรวมโปรเจกต์
CashLog API เป็นบริการหลังบ้านเขียนด้วยภาษา Go ที่ออกแบบมาเพื่อจัดการบันทึกรายรับรายจ่ายจากข้อมูลสลิป พร้อมฟีเจอร์สรุปยอดรายเดือนและรายปี โดยเชื่อมต่อกับฐานข้อมูล PostgreSQL, แคชด้วย Redis และใช้งาน Gemini AI เพื่ออ่านข้อมูลจากรูปภาพสลิปอัตโนมัติ

## ฟีเจอร์หลัก
- อัปโหลดภาพสลิปแล้วให้ระบบอ่านข้อมูลอัตโนมัติด้วย Gemini AI
- บันทึกธุรกรรมรายรับ/รายจ่ายลง PostgreSQL
- ดึงประวัติรายการธุรกรรมตามเดือน/ปี
- สรุปยอดรายรับและรายจ่ายรายเดือนหรือรายปี
- จัดการหมวดหมู่ (categories) ของรายรับและรายจ่าย
- แก้ไขและลบธุรกรรม
- ระบบแคชสรุปแดชบอร์ดด้วย Redis เพื่อลดโหลดฐานข้อมูล
- Health check และ metrics endpoint สำหรับตรวจสอบสถานะระบบ

## โครงสร้างโฟลเดอร์
- `cmd/api/` - entry point ของแอปพลิเคชันและการเชื่อมต่อ dependency
- `cmd/config/` - โหลดค่าการตั้งค่าจาก environment / `.env`
- `internal/delivery/http/` - HTTP delivery layer ของ Fiber
  - `router/` - ตั้งค่า route และ global error handler
  - `middleware/` - middleware สำหรับ CORS, logger, recover, timezone
  - `v1/handler/` - HTTP handlers สำหรับ transaction, category และ health
  - `v1/dto/` - DTO สำหรับ request payload validation
- `internal/domain/` - entity และ interface ของ domain layer
- `internal/usecase/` - business logic ของ transaction และ category
- `internal/repository/postgres/` - repository implementation สำหรับ PostgreSQL ด้วย GORM
- `internal/repository/redis/` - repository implementation สำหรับ Redis cache
- `internal/infrastructure/gemini/` - Gemini AI client และสลิป scanner
- `pkg/database/` - init database, migrate และ seed data
- `pkg/logger/` - wrapper สำหรับ Zap logger
- `pkg/timeutil/` - helper ฟังก์ชันวันที่/เวลา
- `pkg/validate/` - wrapper สำหรับ validation
- `docker-compose.yml` - คอนฟิก docker compose สำหรับรัน API และ Redis
- `Dockerfile` - multi-stage build สำหรับ container production image

## วิธีการติดตั้งและรัน (Setup & Installation)
### 1. เตรียม environment
สร้างไฟล์ `.env` ที่ root โปรเจกต์ แล้วกำหนดค่าดังนี้:

```env
APP_ENV=development
PORT=8080
DB_URL=postgres://user:password@host:5432/dbname?sslmode=disable
DB_MAX_IDEL_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=30m
REDIS_HOST=localhost:6379
REDIS_USERNAME=default
REDIS_PASSWORD=
REDIS_DB=0
GOOGLE_CLOUD_PROJECT=your-gcp-project-id
GEMINI_API_KEY=your-gemini-api-key
MODEL_NAME=gemini-2.5-flash
```

ถ้าต้องการใช้งาน Gemini API ผ่าน Google Cloud ให้เตรียมไฟล์ `google-credentials.json` และวางไว้ที่ root โปรเจกต์ด้วย

### 2. รันด้วย Docker Compose
โปรเจกต์มี `docker-compose.yml` สำหรับรัน API พร้อม Redis cache

```bash
docker compose up --build
```

หลังสั่งรันแล้ว API จะพร้อมใช้งานที่ `http://localhost:8080`

### 3. รันด้วย Go โดยตรง
ถ้าต้องการรันด้วย Go local ให้ติดตั้ง dependency และสั่งรัน

```bash
go mod download
go run ./cmd/api/main.go
```

## Endpoints ที่สำคัญ
- `GET /health` - ตรวจสอบสถานะระบบ
- `GET /metrics` - ดู metrics ของ Fiber app
- `GET /api/v1/transactions` - ดึงประวัติรายการธุรกรรมตามเดือน/ปี
- `GET /api/v1/summary` - ดูสรุปยอดรายรับ/รายจ่ายรายเดือนหรือรายปี
- `POST /api/v1/upload-slip` - อัปโหลดภาพสลิปเพื่อบันทึกรายการ
- `PATCH /api/v1/transactions/:id` - แก้ไขรายการธุรกรรม
- `DELETE /api/v1/transactions/:id` - ลบรายการธุรกรรม
- `POST /api/v1/categories` - สร้างหมวดหมู่ใหม่
- `GET /api/v1/categories?type=expense` - ดึงหมวดหมู่ตามประเภท `expense` หรือ `income`
- `PATCH /api/v1/categories/:id` - แก้ไขหมวดหมู่
- `DELETE /api/v1/categories/:id` - ลบหมวดหมู่

## เทคโนโลยีที่ใช้ (Tech Stack)
- ภาษา: Go
- Web framework: Fiber
- ORM: GORM
- Database: PostgreSQL
- Cache: Redis
- AI OCR / NLP: Gemini AI (Google GenAI)
- Logger: Zap
- Validation: go-playground/validator
- Container: Docker / Docker Compose
- Tools: dotenv, Redis Go client, Google Gemini SDK
