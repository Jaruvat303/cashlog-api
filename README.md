# 💰 CashLog API

**A production-style REST API for automated expense tracking — built with Go, Fiber, and Clean Architecture.**

CashLog API reads receipt/slip images with **Gemini AI OCR**, records income & expense transactions in PostgreSQL, and serves cached monthly/yearly summaries via Redis. Designed as a demonstration of clean, testable, and production-minded API design in Go.

<p align="left">
  <img src="https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go&logoColor=white" alt="Go version">
  <img src="https://img.shields.io/badge/Fiber-v2-black?logo=fiber&logoColor=white" alt="Fiber">
  <img src="https://img.shields.io/badge/PostgreSQL-GORM-336791?logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/Redis-Cache-DC382D?logo=redis&logoColor=white" alt="Redis">
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker&logoColor=white" alt="Docker">
  <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
</p>

> 🇹🇭 อ่านเวอร์ชันภาษาไทยได้ที่ [README.th.md](./README.th.md)

---

## Table of Contents

- [Overview](#overview)
- [Key Features](#key-features)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Reference](#api-reference)
- [Design Decisions](#design-decisions)
- [Roadmap](#roadmap)
- [License](#license)

---

## Overview

Most expense-tracker demos stop at basic CRUD. CashLog API goes further by combining:

- **AI-powered data entry** — upload a photo of a receipt, Gemini AI extracts the amount, date, and category automatically, removing manual entry.
- **Clean Architecture** — strict separation between delivery (HTTP/Fiber), business logic (usecase), and data access (repository), making the codebase easy to test and extend.
- **Performance-conscious design** — dashboard summaries are cached in Redis to reduce repeated aggregation queries against PostgreSQL.

The goal of this project is to demonstrate how I approach designing and structuring a real-world backend service, not just make it "work."

## Key Features

| Feature | Description |
|---|---|
| 🧾 Slip OCR upload | Upload a receipt image and Gemini AI extracts transaction data automatically |
| 💵 Transaction management | Create, list, update, and delete income/expense records |
| 📊 Monthly/yearly summaries | Aggregated income & expense reports, cached in Redis |
| 🏷️ Category management | Custom categories for income and expense types |
| ⚡ Redis-backed caching | Dashboard summary caching to reduce DB load |
| 🩺 Health & metrics endpoints | `/health` and `/metrics` for observability |
| 🐳 Dockerized | Single-command startup via Docker Compose |

## Architecture

CashLog API follows **Clean Architecture**: dependencies point inward, business rules stay independent of frameworks and infrastructure.

```mermaid
flowchart LR
    subgraph Delivery["Delivery Layer (Fiber)"]
        A[HTTP Handlers] --> B[Middleware<br/>CORS · Logger · Recover]
    end

    subgraph Usecase["Usecase Layer"]
        C[Transaction Usecase]
        D[Category Usecase]
    end

    subgraph Domain["Domain Layer"]
        E[Entities & Interfaces]
    end

    subgraph Repository["Repository Layer"]
        F[(PostgreSQL<br/>via GORM)]
        G[(Redis Cache)]
        H[Gemini AI Client]
    end

    A --> C
    A --> D
    C --> E
    D --> E
    E --> F
    E --> G
    C --> H
```

**Request flow example — uploading a slip:**

`POST /api/v1/upload-slip` → Handler validates the request → Usecase calls the Gemini AI client to extract structured data from the image → Usecase persists the transaction via the PostgreSQL repository → Redis summary cache is invalidated → standardized JSON response returned to the client.

## Tech Stack

- **Language:** Go
- **Web Framework:** Fiber
- **ORM:** GORM
- **Database:** PostgreSQL
- **Cache:** Redis
- **AI OCR:** Gemini AI (Google GenAI)
- **Logging:** Zap
- **Validation:** go-playground/validator
- **Containerization:** Docker / Docker Compose
- **CI:** GitHub Actions

## Project Structure

```
cmd/
  api/            # application entry point & dependency wiring
  config/         # environment / .env configuration loading
internal/
  delivery/http/
    router/       # route setup & global error handler
    middleware/   # CORS, logger, recover, timezone
    v1/handler/   # HTTP handlers (transaction, category, health)
    v1/dto/       # request/response DTOs & validation
  domain/         # entities & interfaces (framework-independent)
  usecase/        # business logic
  repository/
    postgres/     # PostgreSQL repository implementation (GORM)
    redis/        # Redis cache repository implementation
  infrastructure/
    gemini/       # Gemini AI client & slip scanner
pkg/
  database/       # DB init, migrations, seed data
  logger/         # Zap logger wrapper
  timeutil/       # date/time helpers
  validate/       # validation wrapper
docker-compose.yml
Dockerfile        # multi-stage production build
```

## Getting Started

### Prerequisites
- Go 1.22+
- Docker & Docker Compose (recommended)
- A Gemini API key ([Google AI Studio](https://ai.google.dev/))

### 1. Configure environment

Create a `.env` file in the project root:

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

If using Gemini via Google Cloud, place a `google-credentials.json` file in the project root as well.

### 2. Run with Docker Compose

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`.

### 3. Run locally with Go

```bash
go mod download
go run ./cmd/api/main.go
```

## API Reference

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/health` | Service health check |
| `GET` | `/metrics` | Fiber app metrics |
| `GET` | `/api/v1/transactions` | List transactions by month/year |
| `POST` | `/api/v1/upload-slip` | Upload a receipt image to auto-create a transaction |
| `PATCH` | `/api/v1/transactions/:id` | Update a transaction |
| `DELETE` | `/api/v1/transactions/:id` | Delete a transaction |
| `GET` | `/api/v1/summary` | Monthly/yearly income & expense summary |
| `POST` | `/api/v1/categories` | Create a category |
| `GET` | `/api/v1/categories?type=expense` | List categories by type (`expense` / `income`) |
| `PATCH` | `/api/v1/categories/:id` | Update a category |
| `DELETE` | `/api/v1/categories/:id` | Delete a category |

**Example — `POST /api/v1/upload-slip`**

```bash
curl -X POST http://localhost:8080/api/v1/upload-slip \
  -H "Content-Type: multipart/form-data" \
  -F "slip=@receipt.jpg"
```

```jsonc
// 200 OK
{
  "success": true,
  "data": {
    "id": 42,
    "type": "expense",
    "amount": 150.00,
    "category": "Food & Beverage",
    "transaction_date": "2026-07-28",
    "note": "Auto-extracted from slip via Gemini AI"
  },
  "meta": {
    "request_id": "b1f2c3-..."
  }
}
```

> ℹ️ Replace this with a real captured example from your own run — recruiters value seeing real output over a hypothetical one.

## Design Decisions

A few notes on *why*, not just *what*:

- **Clean Architecture over a simpler MVC layout** — keeps business logic (usecase) fully decoupled from Fiber and GORM, so the core logic can be unit-tested without spinning up a real database or HTTP server.
- **Redis caching on the summary endpoint** — monthly/yearly aggregation queries are the most expensive reads in the system; caching them avoids recomputation on every dashboard load while transaction writes invalidate the relevant cache keys.
- **Centralized `GlobalErrorHandler`** — maps domain/validation errors to consistent HTTP status codes and a standardized `SuccessResponseDTO` / `ErrorResponseDTO` shape, so API consumers always get a predictable response contract.
- **DTO layer separate from domain entities** — request/response shapes can evolve independently of internal data models, and validation stays isolated at the edge of the system.

## Roadmap

- [ ] Unit tests for usecase layer (mocked repositories)
- [ ] OpenAPI/Swagger documentation
- [ ] JWT-based authentication
- [ ] CI badge linked to GitHub Actions pipeline
- [ ] Live demo deployment

## License

This project is licensed under the [MIT License](./LICENSE).

---

<p align="center">Built by <a href="https://github.com/Jaruvat303">Jaruvat303</a> — feedback and PRs welcome.</p>
