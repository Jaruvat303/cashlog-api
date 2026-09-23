# Spec: Income vs Expense Trend API (backend)

> คู่กับ spec ฝั่งแอป: `cashlog-app` → planning/specs/income-expense-bar-chart-page.md
> ที่มา: grill-me session ก.ย. 2026 (Q1–Q16)

## Problem Statement

ตอนนี้ Jar ดูรายรับรายจ่ายได้ทีละเดือนผ่าน `GET /api/v1/transactions/summary` เท่านั้น ถ้าอยากเห็นแนวโน้มทั้งปี (เดือนไหนใช้เงินเกินรายรับ) หรือเทียบปีต่อปี แอปจะต้องยิง summary 12 ครั้งต่อการเปิดหน้าเดียว ซึ่งกิน rate limit (60 req/60s) ไปราว 20% และดึง category breakdown ที่ไม่ได้ใช้มาด้วย นอกจากนี้ตาราง `transactions` ยังไม่มี index บน `transaction_date` เลย ทั้ง summary เดิมและ query ใหม่จึงต้อง scan ทั้งตาราง

## Solution

เพิ่ม endpoint ใหม่ `GET /api/v1/transactions/trend` ที่ aggregate รายรับ/รายจ่ายด้วย SQL ครั้งเดียวแล้วคืนเป็นชุด bucket ที่พร้อมวาดกราฟแท่งทันที รองรับสองโหมด:

- **month**: 12 bucket (ม.ค.–ธ.ค.) ของปีที่เลือก ครบเสมอ เดือนว่างเป็น 0
- **year**: 1 bucket ต่อปี ตั้งแต่ปีแรกที่มีรายการถึงปีปัจจุบัน (สูงสุด 10 ปี)

ขอบเดือน/ปีคิดตามเวลา `Asia/Bangkok` ด้วย helper ตัวเดียวกับ `CalculateSummary` เพื่อให้ยอดของเดือนเดียวกันตรงกันทั้งสอง endpoint พร้อมเพิ่ม index บน `transaction_date`

## User Stories

1. As Jar, I want to fetch income and expense totals for all 12 months of a year in one request, so that the app can draw a monthly bar chart without spending my rate-limit budget.
2. As Jar, I want to fetch yearly income and expense totals across every year I've used the app, so that I can compare year over year.
3. As Jar, I want months with no transactions to come back as zero rather than missing, so that the chart's X axis is always complete and the app has no calendar logic.
4. As Jar, I want future months of the current year to be included as zero buckets, so that the response shape is always exactly 12 buckets in month mode.
5. As Jar, I want each bucket to include `net` (income − expense), so that its definition matches the summary endpoint exactly.
6. As Jar, I want transfers excluded from both totals, so that moving money between my own accounts never appears as income or expense.
7. As Jar, I want a transaction made at 00:30 Bangkok time on the 1st to count in that month, so that month boundaries match what I see in real life.
8. As Jar, I want the March total in the trend endpoint to equal the March total from the summary endpoint, so that the two screens never contradict each other.
9. As Jar, I want the month mode to default to the current year when I omit `year`, so that the common case needs no parameters.
10. As Jar, I want `granularity` to default to `month`, so that the simplest call returns the most useful view.
11. As Jar, I want an invalid `granularity` to return 400 with a field-level error, so that client bugs surface clearly.
12. As Jar, I want a `year` outside 2000 to (current year + 1) to return 400 with a field-level error, so that nonsense queries are rejected.
13. As Jar, I want year mode to start from the first year that has any transaction, so that I don't see empty bars for years before I used the app.
14. As Jar, I want year mode capped at the 10 most recent years, so that the response stays bounded.
15. As Jar, I want year mode on an empty database to return one zero bucket for the current year, so that the app never has to special-case an empty array.
16. As Jar, I want new or edited transactions to appear in the trend immediately, so that the chart is never stale after scanning a slip.
17. As Jar, I want the trend query and the existing summary query to use an index on `transaction_date`, so that both stay fast as data grows.
18. As Jar, I want the endpoint documented in swagger, so that the app side can verify the contract against dev before coding.
19. As Jar, I want the response wrapped in the same success envelope every other endpoint uses, so that the app's existing `ApiClient` parses it without changes.

## Implementation Decisions

**API contract**

- `GET /api/v1/transactions/trend`
- Query: `granularity` = `month` | `year` (default `month`); `year` = int, used only in month mode (default current Bangkok year)
- Response (inside `pkg/response.JsonResponse.data`, via `response.Success` — the same envelope every other endpoint uses: `success` / `data` / `message`, no `request_id` or `timestamp`):

```
TrendResponse {
  granularity: "month" | "year"
  year: int | null            // present only in month mode
  buckets: [
    { year: int, month: int | null, total_income: number, total_expense: number, net: number }
  ]                           // month = null in year mode
}
```

- Field names `total_income` / `total_expense` match `DashboardSummaryResponse`
- Month is a pointer type in Go so it serialises as JSON `null`
- Buckets are sorted in ascending chronological order

**Domain / ports**

- New domain type `TrendBucket` (year, optional month, totals, net) and a granularity enum
- `domain.TransactionRepository` gains two methods:
  - `AggregateMonthly(from, to)` — always aggregates income/expense totals grouped by **month** (Bangkok time) within a time range. Granularity never reaches the repository or SQL; the usecase rolls months up into years itself for year mode.
  - `GetFirstTransactionYear()` — returns the first year that has any transaction (or none)
- No new usecase: `TransactionUseCase` gains `GetTrend(granularity, year)` — this keeps the rule of no usecase-to-usecase dependencies

**Shared period helper (prefactor)**

- Extract the logic that turns (year, month) into a Bangkok-local `[start, end)` range as `timestamptz` into one helper
- Refactor `CalculateSummary` to use it before trend is built, so both endpoints share one definition of "a month"

**Usecase behaviour (`GetTrend`)**

- Validates input; an invalid `granularity` or out-of-range `year` returns `domain.ErrInvalidInput` wrapped with a message naming the offending field (e.g. `"invalid input parameters: year must be between 2000 and 2027, got 1999"`) — `GlobalErrorHandler` already maps `ErrInvalidInput` to 400 `INVALID_INPUT_PARAMETERS`; there is no separate field-level error structure in this codebase
- Month mode: range = whole Bangkok year; builds 12 buckets first, then merges `AggregateMonthly` results into them
- Year mode: range = Jan 1 of max(first year, current year − 9) to end of current year; builds one bucket per year, then rolls the monthly aggregates up and merges; empty DB → single zero bucket for current year; a first-transaction-year in the future (bad data) clamps to the current year
- Computes `net` = income − expense per bucket; `net` and any rolled-up yearly totals are rounded to 2 decimals (`math.Round(x*100)/100`) to absorb float64 summation drift
- Compares `transaction_type` using the domain's lowercase constants (avoid repeating the old uppercase `"INCOME"` bug in `CalculateSummary`)
- Transfers are excluded

**Repository query**

- `AggregateMonthly` always groups by month — the usecase is the only place granularity is known; it rolls months up into years itself for year mode
- Filter must be sargable so the index is used: `transaction_date >= from AND transaction_date < to`
- Grouping uses the truncated Bangkok local time of `transaction_date`
- Sums income and expense conditionally in the same query, with `income`/`expense` passed in as query args from the domain constants (not SQL literals)
- Built off `.Model(&domain.Transaction{})` (like `CalculateSummary`), not a hardcoded `"transactions"` table string — dev/prod schema separation here is via `search_path` on the DSN, not a GORM `NamingStrategy.TablePrefix`, but resolving the table name through the model keeps the query correct if that ever changes
- `GetFirstTransactionYear` wraps `MIN(transaction_date)` in `AT TIME ZONE 'Asia/Bangkok'` *before* extracting the year (`EXTRACT(YEAR FROM (MIN(transaction_date) AT TIME ZONE 'Asia/Bangkok'))`), so `MIN` itself can still use `idx_transactions_transaction_date`

**Schema**

- Add a GORM `index` tag on `TransactionDate` so AutoMigrate creates the index on dev and prod
- Verified from the dev schema: the table currently has only the primary key; FKs do not create indexes automatically in Postgres

**Caching**

- No Redis cache for this endpoint (single-user data, cheap query); avoids needing invalidation on every create/update/delete

**Delivery**

- New handler on the transactions route group, responding via `response.Success` (the helper every other endpoint uses — there is no `response.OK`), with swagger annotations; swagger is regenerated into `docs/` (not hand-edited)

## Testing Decisions

- A good test hits external behaviour only: given transactions in the repository, what buckets come out. It does not assert on how the SQL is built or on internal helper calls.
- **Primary seam: `TransactionUseCase.GetTrend`** with a fake `TransactionRepository` port. Covers:
  - month mode always returns 12 buckets, in order, empty months zero-filled
  - future months of the current year are zero buckets
  - `net` is correct; transfers do not affect totals
  - year mode starts at the first transaction year; capped at 10 years; empty DB → one zero bucket
  - default granularity and default year
  - invalid granularity / out-of-range year → validation error
- **Period helper**: direct tests for the Bangkok boundary — a transaction at 00:30 on the 1st (Bangkok) belongs to that month, while 23:30 on the last day belongs to that month too; also the Dec → Jan year rollover
- **`CalculateSummary` regression**: its existing tests must keep passing after the refactor onto the shared helper
- **Repository SQL**: the fake port cannot prove the timezone truncation in SQL, so verify on the dev Cloud Run instance against known data (a transaction near a month boundary) and confirm the March trend total equals the March summary total
- Prior art: existing unit tests of `TransactionUseCase` and `CalculateSummary`

## Out of Scope

- Redis caching of trend results
- Per-category breakdown in the trend (the pie chart on the summary tab already covers categories)
- Custom date ranges (`from_year` / `to_year`) or a rolling 12-month window across years
- Showing transfers as a third series
- Any frontend work (see the cashlog-app spec)

## Further Notes

- The app has data only from July 2026, so year mode will show a single year for a while; this is expected.
- The index on `transaction_date` also speeds up the existing summary endpoint.
- Merge flow: feature branch → develop (verify on dev Cloud Run) → main. The app's data-layer ticket depends on this endpoint being live on dev.
