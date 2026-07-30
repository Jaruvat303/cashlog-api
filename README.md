# 💰 CashLog API

<p align="center">
  <img alt="Go Version" src="https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go">
  <img alt="Framework" src="https://img.shields.io/badge/Framework-Fiber-00ACD7?style=flat-square">
  <img alt="License" src="https://img.shields.io/badge/License-MIT-yellow.svg?style=flat-square">
  <img alt="Build" src="https://img.shields.io/github/actions/workflow/status/Jaruvat303/cashlog-api/ci.yml?branch=main&style=flat-square&label=CI%2FCD">
  <img alt="Docker" src="https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker">
</p>

<p align="center">
  <b>Backend API สำหรับบันทึกรายรับ-รายจ่าย พร้อมระบบอ่านสลิปอัตโนมัติด้วย AI</b><br>
  เขียนด้วย Go ตามหลัก Clean Architecture — ออกแบบมาให้ maintain และ scale ได้จริงในระดับ production
</p>

<p align="center">
  🔗 <a href="https://cashlog-api-dev-501276653507.asia-southeast1.run.app/api/v1/transactions">Live Demo</a> •
  📄 <a href="https://cashlog-api-dev-501276653507.asia-southeast1.run.app/swagger/index.html">API Documentation (Swagger)</a> •
  🐳 <a href="#การติดตั้งและรัน">Setup Guide</a>
</p>

---

## ✨ จุดเด่นของโปรเจกต์

| | |
|---|---|
| 🏗️ **Clean Architecture** | แยก layer ชัดเจน (delivery / usecase / domain / repository) พร้อม standardized response DTO และ centralized error handler ทำให้โค้ด testable และขยายต่อได้ง่าย |
| 🤖 **AI-Powered OCR** | ใช้ Gemini AI อ่านข้อมูลจากภาพสลิปโดยอัตโนมัติ ลดการกรอกข้อมูลด้วยมือ |
| ⚙️ **Production-Grade DevOps** | มี CI/CD pipeline ผ่าน GitHub Actions, multi-stage Dockerfile, pre-commit hooks และ unit test ครอบคลุม business logic หลัก |

---

## 📐 สถาปัตยกรรมระบบ

```mermaid
flowchart TD
    Client["Client / Postman / Frontend"]

    subgraph API["CashLog API (Go + Fiber)"]
        Router["Router + Middleware<br/>(CORS, Logger, Recover, Timezone)"]
        Handler["HTTP Handler (v1)"]
        Usecase["Usecase Layer<br/>(Business Logic)"]
        RepoPG["Repository<br/>(PostgreSQL / GORM)"]
        RepoRedis["Repository<br/>(Redis Cache)"]
        Gemini["Gemini Client<br/>(Slip Scanner)"]
    end

    DB[("PostgreSQL")]
    Cache[("Redis")]
    AI["Gemini AI API"]

    Client --> Router --> Handler --> Usecase
    Usecase --> RepoPG --> DB
    Usecase --> RepoRedis --> Cache
    Usecase --> Gemini --> AI
```

### Flow: อัปโหลดสลิป → บันทึกธุรกรรมอัตโนมัติ

```mermaid
sequenceDiagram
    participant U as User
    participant H as Handler
    participant G as Gemini Client
    participant UC as Usecase
    participant DB as PostgreSQL
    participant R as Redis

    U->>H: POST /api/v1/upload-slip (image)
    H->>G: ส่งภาพสลิปให้ Gemini อ่านข้อมูล
    G-->>H: คืนข้อมูล (จำนวนเงิน, วันที่, หมวดหมู่)
    H->>UC: สร้าง Transaction จากข้อมูลที่อ่านได้
    UC->>DB: บันทึกธุรกรรม
    UC->>R: Invalidate cache สรุปแดชบอร์ด
    UC-->>U: ตอบกลับ Transaction ที่บันทึกสำเร็จ
```

---

## 🚀 ฟีเจอร์หลัก

- 📷 อัปโหลดภาพสลิปแล้วให้ระบบอ่านข้อมูลอัตโนมัติด้วย Gemini AI
- 💵 บันทึกธุรกรรมรายรับ/รายจ่ายลง PostgreSQL
- 📅 ดึงประวัติรายการธุรกรรมตามเดือน/ปี
- 📊 สรุปยอดรายรับและรายจ่ายรายเดือนหรือรายปี
- 🏷️ จัดการหมวดหมู่ (categories) ของรายรับและรายจ่าย
- ✏️ แก้ไขและลบธุรกรรม
- ⚡ ระบบแคชสรุปแดชบอร์ดด้วย Redis เพื่อลดโหลดฐานข้อมูล
- ❤️ Health check และ metrics endpoint สำหรับตรวจสอบสถานะระบบ

---

## 🛠️ เทคโนโลยีที่ใช้

| Layer | Technology |
|---|---|
| Language | Go |
| Web Framework | Fiber |
| ORM | GORM |
| Database | PostgreSQL |
| Cache | Redis |
| AI OCR | Gemini AI (Google GenAI) |
| Logger | Zap |
| Validation | go-playground/validator |
| API Docs | Swagger / OpenAPI |
| Container | Docker / Docker Compose |
| CI/CD | GitHub Actions |
| Testing | Go testing + testify |

---

## 📂 โครงสร้างโฟลเดอร์

```
cmd/
├── api/             entry point ของแอปพลิเคชันและการเชื่อมต่อ dependency
└── config/          โหลดค่าการตั้งค่าจาก environment / .env

internal/
├── delivery/http/
│   ├── router/      route และ global error handler
│   ├── middleware/   CORS, logger, recover, timezone
│   └── v1/
│       ├── handler/  HTTP handlers (transaction, category, health)
│       └── dto/      DTO สำหรับ request payload validation
├── domain/          entity และ interface ของ domain layer
├── usecase/         business logic (transaction, category)
├── repository/
│   ├── postgres/    repository implementation ด้วย GORM
│   └── redis/       repository implementation สำหรับ Redis cache
└── infrastructure/
    └── gemini/      Gemini AI client และสลิป scanner

pkg/
├── database/        init database, migrate, seed data
├── logger/          wrapper สำหรับ Zap logger
├── timeutil/        helper ฟังก์ชันวันที่/เวลา
└── validate/        wrapper สำหรับ validation

docker-compose.yml    รัน API + Redis
Dockerfile            multi-stage build สำหรับ production image
```

---

## ⚙️ การติดตั้งและรัน

### 1. เตรียม environment

สร้างไฟล์ `.env` ที่ root โปรเจกต์:

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

ถ้าต้องการใช้งาน Gemini API ผ่าน Google Cloud ให้เตรียมไฟล์ `google-credentials.json` วางไว้ที่ root โปรเจกต์ด้วย

### 2. รันด้วย Docker Compose (แนะนำ)

```bash
docker compose up --build
```

API จะพร้อมใช้งานที่ `http://localhost:8080`

### 3. รันด้วย Go โดยตรง

```bash
go mod download
go run ./cmd/api/main.go
```

### 4. รัน tests

```bash
go test ./... -v -cover
```

---

## 📡 API Endpoints

| Method | Endpoint | คำอธิบาย |
|---|---|---|
| GET | `/health` | ตรวจสอบสถานะระบบ |
| GET | `/metrics` | ดู metrics ของ Fiber app |
| GET | `/api/v1/transactions` | ดึงประวัติรายการธุรกรรมตามเดือน/ปี |
| GET | `/api/v1/summary` | ดูสรุปยอดรายรับ/รายจ่ายรายเดือนหรือรายปี |
| POST | `/api/v1/upload-slip` | อัปโหลดภาพสลิปเพื่อบันทึกรายการ |
| PATCH | `/api/v1/transactions/:id` | แก้ไขรายการธุรกรรม |
| DELETE | `/api/v1/transactions/:id` | ลบรายการธุรกรรม |
| POST | `/api/v1/categories` | สร้างหมวดหมู่ใหม่ |
| GET | `/api/v1/categories?type=expense` | ดึงหมวดหมู่ตามประเภท `expense` หรือ `income` |
| PATCH | `/api/v1/categories/:id` | แก้ไขหมวดหมู่ |
| DELETE | `/api/v1/categories/:id` | ลบหมวดหมู่ |

📄 API spec แบบเต็มดูได้ที่ [Swagger UI](https://YOUR-DEPLOY-URL.example.com/swagger/index.html)

---

## 🔄 CI/CD Pipeline

โปรเจกต์นี้มี GitHub Actions pipeline ที่รันอัตโนมัติทุกครั้งที่ push/PR ไปที่ `main`:

- ✅ Lint & format check
- ✅ Unit test + coverage report
- ✅ Build Docker image (multi-stage)
- ✅ Deploy อัตโนมัติเมื่อผ่านทุกขั้นตอน

ดู workflow ได้ที่ [`.github/workflows`](.github/workflows)

---

## 📝 License

โปรเจกต์นี้เผยแพร่ภายใต้ [MIT License](LICENSE)

---

## 👤 ผู้พัฒนา

**Jaruvat** — Senior API Designer
📫 GitHub: [@Jaruvat303](https://github.com/Jaruvat303)
