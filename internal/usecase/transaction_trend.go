package usecase

import (
	"context"
	"fmt"
	"math"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
)

const trendMinYear = 2000

// yearlyTrendCapYears cือจำนวนปีล่าสุดสูงสุดที่ granularity=year จะแสดงย้อนหลัง (Decision spec: สูงสุด 10 ปี)
const yearlyTrendCapYears = 10

// roundMoney ปัดเศษเงินเป็น 2 ตำแหน่งทศนิยม เพื่อลบล้าง float64 drift ที่เกิดจากการบวกลบยอดหลายรายการ
func roundMoney(amount float64) float64 {
	return math.Round(amount*100) / 100
}

// GetTrend implements [domain.TransactionUsecase]. คืนชุด bucket รายรับ-รายจ่ายสำหรับกราฟแท่ง (Ticket B2)
func (t *transactionUsecase) GetTrend(ctx context.Context, granularity string, year int) (*domain.TrendResult, error) {
	if granularity == "" {
		granularity = domain.TrendGranularityMonth
	}
	if granularity != domain.TrendGranularityMonth && granularity != domain.TrendGranularityYear {
		return nil, fmt.Errorf("%w: granularity must be 'month' or 'year', got '%s'", domain.ErrInvalidInput, granularity)
	}

	currentYear := timeutil.NowInBangkok().Year()

	if granularity == domain.TrendGranularityYear {
		return t.getYearlyTrend(ctx, currentYear)
	}

	// year ใช้เฉพาะ month mode เท่านั้น (API contract) — 0 แปลว่า client ไม่ได้ส่งมา ใช้ปีปัจจุบันแทน
	if year == 0 {
		year = currentYear
	}
	if year < trendMinYear || year > currentYear+1 {
		return nil, fmt.Errorf("%w: year must be between %d and %d, got %d", domain.ErrInvalidInput, trendMinYear, currentYear+1, year)
	}

	return t.getMonthlyTrend(ctx, year)
}

// getMonthlyTrend คืน 12 bucket ของปีที่ระบุเสมอ (เดือนที่ไม่มีข้อมูล รวมถึงเดือนในอนาคต จะเป็นศูนย์)
func (t *transactionUsecase) getMonthlyTrend(ctx context.Context, year int) (*domain.TrendResult, error) {
	from, to := timeutil.YearRangeBangkok(year)

	aggregates, err := t.txRepo.AggregateMonthly(ctx, from, to)
	if err != nil {
		return nil, err
	}

	byMonth := make(map[int]domain.MonthlyAggregate, len(aggregates))
	for _, agg := range aggregates {
		byMonth[agg.Month] = agg
	}

	// สร้าง 12 bucket ก่อน (zero-fill) แล้วค่อย merge ผลลัพธ์จาก aggregate ทับเข้าไป
	buckets := make([]domain.TrendBucket, 12)
	for i := range buckets {
		month := i + 1
		bucket := domain.TrendBucket{Year: year, Month: &month}

		if agg, ok := byMonth[month]; ok {
			bucket.TotalIncome = agg.TotalIncome
			bucket.TotalExpense = agg.TotalExpense
		}
		bucket.Net = roundMoney(bucket.TotalIncome - bucket.TotalExpense)

		buckets[i] = bucket
	}

	resultYear := year
	return &domain.TrendResult{
		Granularity: domain.TrendGranularityMonth,
		Year:        &resultYear,
		Buckets:     buckets,
	}, nil
}

// getYearlyTrend คืน 1 bucket ต่อปี ตั้งแต่ max(ปีแรกที่มีรายการ, ปีปัจจุบัน-9) ถึงปีปัจจุบัน
// ฐานข้อมูลว่าง -> bucket เดียวของปีปัจจุบัน, ค่าเป็นศูนย์ทั้งหมด
func (t *transactionUsecase) getYearlyTrend(ctx context.Context, currentYear int) (*domain.TrendResult, error) {
	firstYear, hasData, err := t.txRepo.GetFirstTransactionYear(ctx)
	if err != nil {
		return nil, err
	}

	if !hasData {
		return &domain.TrendResult{
			Granularity: domain.TrendGranularityYear,
			Buckets:     []domain.TrendBucket{{Year: currentYear}},
		}, nil
	}

	// กันเคสข้อมูลเพี้ยน (transaction วันที่ในอนาคต) ทำให้ปีแรกเลยปีปัจจุบันไปได้
	if firstYear > currentYear {
		firstYear = currentYear
	}

	startYear := firstYear
	if capStart := currentYear - (yearlyTrendCapYears - 1); startYear < capStart {
		startYear = capStart
	}

	// สร้าง bucket ต่อปีก่อน (zero-fill) แล้วค่อย roll up ผลรายเดือนจาก repo ทับเข้าไป
	buckets := make([]domain.TrendBucket, currentYear-startYear+1)
	byYear := make(map[int]*domain.TrendBucket, len(buckets))
	for i := range buckets {
		buckets[i] = domain.TrendBucket{Year: startYear + i}
		byYear[buckets[i].Year] = &buckets[i]
	}

	from, _ := timeutil.YearRangeBangkok(startYear)
	_, to := timeutil.YearRangeBangkok(currentYear)

	aggregates, err := t.txRepo.AggregateMonthly(ctx, from, to)
	if err != nil {
		return nil, err
	}

	for _, agg := range aggregates {
		bucket, ok := byYear[agg.Year]
		if !ok {
			continue
		}
		bucket.TotalIncome += agg.TotalIncome
		bucket.TotalExpense += agg.TotalExpense
	}

	for i := range buckets {
		buckets[i].TotalIncome = roundMoney(buckets[i].TotalIncome)
		buckets[i].TotalExpense = roundMoney(buckets[i].TotalExpense)
		buckets[i].Net = roundMoney(buckets[i].TotalIncome - buckets[i].TotalExpense)
	}

	return &domain.TrendResult{
		Granularity: domain.TrendGranularityYear,
		Buckets:     buckets,
	}, nil
}
