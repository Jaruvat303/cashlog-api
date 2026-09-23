package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/internal/usecase"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// newTrendUsecase สร้าง TransactionUsecase พร้อม fake repository port สำหรับเทส GetTrend โดยเฉพาะ
// (ไม่ต้องพึ่ง cache/gemini/account/category repo เพราะ GetTrend ไม่แตะ dependency พวกนั้นเลย)
func newTrendUsecase(repo *domain.TransactionRepositoryMock) domain.TransactionUsecase {
	return usecase.NewTransactionUsecase(repo, nil, nil, nil, nil, nil, logger.NewNopLogger())
}

func intPtr(i int) *int { return &i }

func TestGetTrend_Month(t *testing.T) {
	now := timeutil.NowInBangkok()
	currentYear := now.Year()

	t.Run("1. always returns 12 ascending buckets, zero-filling months with no data", func(t *testing.T) {
		year := currentYear - 1 // ปีที่ผ่านไปแล้วทั้งปี ไม่ต้องกังวลเรื่องเดือนอนาคต
		from, to := timeutil.YearRangeBangkok(year)

		repo := new(domain.TransactionRepositoryMock)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return([]domain.MonthlyAggregate{
			{Year: year, Month: 1, TotalIncome: 15000, TotalExpense: 5000},
			{Year: year, Month: 6, TotalIncome: 2000, TotalExpense: 1800},
			{Year: year, Month: 12, TotalIncome: 500, TotalExpense: 0},
		}, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityMonth, year)

		assert.NoError(t, err)
		assert.Equal(t, domain.TrendGranularityMonth, result.Granularity)
		assert.Equal(t, &year, result.Year)
		assert.Len(t, result.Buckets, 12)

		for i, bucket := range result.Buckets {
			expectedMonth := i + 1
			assert.Equal(t, year, bucket.Year)
			assert.Equal(t, intPtr(expectedMonth), bucket.Month)
		}

		assert.Equal(t, 15000.0, result.Buckets[0].TotalIncome)
		assert.Equal(t, 5000.0, result.Buckets[0].TotalExpense)
		assert.Equal(t, 10000.0, result.Buckets[0].Net)

		// เดือนที่ไม่มีข้อมูลจาก repo ต้องเป็นศูนย์ ไม่ใช่หายไปจากลิสต์
		assert.Equal(t, 0.0, result.Buckets[1].TotalIncome)
		assert.Equal(t, 0.0, result.Buckets[1].TotalExpense)
		assert.Equal(t, 0.0, result.Buckets[1].Net)

		assert.Equal(t, 500.0, result.Buckets[11].TotalIncome)
		assert.Equal(t, 500.0, result.Buckets[11].Net)

		repo.AssertExpectations(t)
	})

	t.Run("2. current year: months with no rows yet (including future months) come back zero", func(t *testing.T) {
		from, to := timeutil.YearRangeBangkok(currentYear)

		repo := new(domain.TransactionRepositoryMock)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return([]domain.MonthlyAggregate{
			{Year: currentYear, Month: 1, TotalIncome: 1000, TotalExpense: 400},
		}, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityMonth, currentYear)

		assert.NoError(t, err)
		assert.Len(t, result.Buckets, 12)
		for i := 1; i < 12; i++ { // เดือน 2-12 ไม่มีข้อมูลจาก repo เลย (รวมเดือนในอนาคต)
			assert.Equal(t, 0.0, result.Buckets[i].TotalIncome)
			assert.Equal(t, 0.0, result.Buckets[i].TotalExpense)
			assert.Equal(t, 0.0, result.Buckets[i].Net)
		}
		repo.AssertExpectations(t)
	})

	t.Run("3. net is rounded to 2 decimals to absorb float64 drift", func(t *testing.T) {
		year := currentYear - 1
		from, to := timeutil.YearRangeBangkok(year)

		repo := new(domain.TransactionRepositoryMock)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return([]domain.MonthlyAggregate{
			{Year: year, Month: 3, TotalIncome: 100.1, TotalExpense: 0.2}, // 100.1-0.2 = 99.90000000000001 ใน float64
		}, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityMonth, year)

		assert.NoError(t, err)
		assert.Equal(t, 99.9, result.Buckets[2].Net)
		repo.AssertExpectations(t)
	})

	t.Run("4. default granularity (empty string) and default year (0) fall back to month/current year", func(t *testing.T) {
		from, to := timeutil.YearRangeBangkok(currentYear)

		repo := new(domain.TransactionRepositoryMock)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return([]domain.MonthlyAggregate{}, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), "", 0)

		assert.NoError(t, err)
		assert.Equal(t, domain.TrendGranularityMonth, result.Granularity)
		assert.Equal(t, &currentYear, result.Year)
		assert.Len(t, result.Buckets, 12)
		repo.AssertExpectations(t)
	})

	t.Run("5. invalid granularity is rejected before touching the repository", func(t *testing.T) {
		repo := new(domain.TransactionRepositoryMock)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), "week", currentYear)

		assert.Nil(t, result)
		assert.True(t, errors.Is(err, domain.ErrInvalidInput))
		repo.AssertExpectations(t) // ไม่มี call ไหนถูกตั้งไว้ -> ต้องไม่มี call เกิดขึ้นจริง
	})

	t.Run("6. year outside [2000, current+1] is rejected before touching the repository", func(t *testing.T) {
		tests := []struct {
			name string
			year int
		}{
			{"below 2000", 1999},
			{"above current+1", currentYear + 2},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo := new(domain.TransactionRepositoryMock)

				u := newTrendUsecase(repo)
				result, err := u.GetTrend(context.Background(), domain.TrendGranularityMonth, tt.year)

				assert.Nil(t, result)
				assert.True(t, errors.Is(err, domain.ErrInvalidInput))
				repo.AssertExpectations(t)
			})
		}
	})

	t.Run("7. repository error propagates unchanged", func(t *testing.T) {
		from, to := timeutil.YearRangeBangkok(currentYear)
		repoErr := errors.New("db exploded")

		repo := new(domain.TransactionRepositoryMock)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return(nil, repoErr)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityMonth, currentYear)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, repoErr)
		repo.AssertExpectations(t)
	})
}

func TestGetTrend_Year(t *testing.T) {
	now := timeutil.NowInBangkok()
	currentYear := now.Year()

	t.Run("1. starts at the first transaction year and rolls monthly data up into yearly buckets", func(t *testing.T) {
		firstYear := currentYear - 3
		from, _ := timeutil.YearRangeBangkok(firstYear)
		_, to := timeutil.YearRangeBangkok(currentYear)

		repo := new(domain.TransactionRepositoryMock)
		repo.On("GetFirstTransactionYear", mock.Anything).Return(firstYear, true, nil)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return([]domain.MonthlyAggregate{
			{Year: firstYear, Month: 1, TotalIncome: 1000, TotalExpense: 200},
			{Year: firstYear, Month: 6, TotalIncome: 500, TotalExpense: 100},
			{Year: currentYear, Month: 1, TotalIncome: 300, TotalExpense: 300},
		}, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityYear, 0)

		assert.NoError(t, err)
		assert.Equal(t, domain.TrendGranularityYear, result.Granularity)
		assert.Nil(t, result.Year)
		assert.Len(t, result.Buckets, 4) // firstYear..currentYear inclusive

		assert.Equal(t, firstYear, result.Buckets[0].Year)
		assert.Nil(t, result.Buckets[0].Month)
		assert.Equal(t, 1500.0, result.Buckets[0].TotalIncome) // 1000+500 rolled up from 2 months
		assert.Equal(t, 300.0, result.Buckets[0].TotalExpense)
		assert.Equal(t, 1200.0, result.Buckets[0].Net)

		assert.Equal(t, firstYear+1, result.Buckets[1].Year)
		assert.Equal(t, 0.0, result.Buckets[1].TotalIncome) // ปีที่ไม่มีข้อมูลเลยต้องเป็นศูนย์

		assert.Equal(t, currentYear, result.Buckets[3].Year)
		assert.Equal(t, 300.0, result.Buckets[3].TotalIncome)
		assert.Equal(t, 0.0, result.Buckets[3].Net)

		repo.AssertExpectations(t)
	})

	t.Run("2. capped at the 10 most recent years", func(t *testing.T) {
		firstYear := currentYear - 15 // เก่ากว่า cap มาก
		cappedStart := currentYear - 9
		from, _ := timeutil.YearRangeBangkok(cappedStart)
		_, to := timeutil.YearRangeBangkok(currentYear)

		repo := new(domain.TransactionRepositoryMock)
		repo.On("GetFirstTransactionYear", mock.Anything).Return(firstYear, true, nil)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return([]domain.MonthlyAggregate{}, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityYear, 0)

		assert.NoError(t, err)
		assert.Len(t, result.Buckets, 10)
		assert.Equal(t, cappedStart, result.Buckets[0].Year)
		assert.Equal(t, currentYear, result.Buckets[9].Year)
		repo.AssertExpectations(t)
	})

	t.Run("3. empty database returns exactly one zero bucket for the current year", func(t *testing.T) {
		repo := new(domain.TransactionRepositoryMock)
		repo.On("GetFirstTransactionYear", mock.Anything).Return(0, false, nil)
		// ไม่ตั้ง AggregateMonthly ไว้เลย -> ถ้า implementation เผลอเรียกจะ panic ทันที (พิสูจน์ว่า short-circuit จริง)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityYear, 0)

		assert.NoError(t, err)
		assert.Len(t, result.Buckets, 1)
		assert.Equal(t, currentYear, result.Buckets[0].Year)
		assert.Nil(t, result.Buckets[0].Month)
		assert.Equal(t, 0.0, result.Buckets[0].TotalIncome)
		assert.Equal(t, 0.0, result.Buckets[0].TotalExpense)
		assert.Equal(t, 0.0, result.Buckets[0].Net)
		repo.AssertExpectations(t)
	})

	t.Run("4. first transaction year in the future clamps to the current year (one bucket)", func(t *testing.T) {
		from, to := timeutil.YearRangeBangkok(currentYear)

		repo := new(domain.TransactionRepositoryMock)
		repo.On("GetFirstTransactionYear", mock.Anything).Return(currentYear+5, true, nil)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return([]domain.MonthlyAggregate{}, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityYear, 0)

		assert.NoError(t, err)
		assert.Len(t, result.Buckets, 1)
		assert.Equal(t, currentYear, result.Buckets[0].Year)
		repo.AssertExpectations(t)
	})

	t.Run("5. yearly totals rolled up from multiple months are rounded to 2 decimals", func(t *testing.T) {
		firstYear := currentYear
		from, to := timeutil.YearRangeBangkok(currentYear)

		repo := new(domain.TransactionRepositoryMock)
		repo.On("GetFirstTransactionYear", mock.Anything).Return(firstYear, true, nil)
		repo.On("AggregateMonthly", mock.Anything, from, to).Return([]domain.MonthlyAggregate{
			{Year: currentYear, Month: 1, TotalIncome: 100.1, TotalExpense: 0},
			{Year: currentYear, Month: 2, TotalIncome: 0, TotalExpense: 0.2},
		}, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityYear, 0)

		assert.NoError(t, err)
		assert.Len(t, result.Buckets, 1)
		assert.Equal(t, 100.1, result.Buckets[0].TotalIncome)
		assert.Equal(t, 0.2, result.Buckets[0].TotalExpense)
		assert.Equal(t, 99.9, result.Buckets[0].Net)
		repo.AssertExpectations(t)
	})

	t.Run("6. GetFirstTransactionYear error propagates unchanged", func(t *testing.T) {
		repoErr := errors.New("db exploded")

		repo := new(domain.TransactionRepositoryMock)
		repo.On("GetFirstTransactionYear", mock.Anything).Return(0, false, repoErr)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityYear, 0)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, repoErr)
		repo.AssertExpectations(t)
	})

	t.Run("7. year query param is ignored in year mode, even when out of range", func(t *testing.T) {
		repo := new(domain.TransactionRepositoryMock)
		repo.On("GetFirstTransactionYear", mock.Anything).Return(0, false, nil)

		u := newTrendUsecase(repo)
		result, err := u.GetTrend(context.Background(), domain.TrendGranularityYear, 1)

		assert.NoError(t, err)
		assert.Len(t, result.Buckets, 1)
		repo.AssertExpectations(t)
	})
}
