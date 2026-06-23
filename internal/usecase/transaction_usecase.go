package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Jaruvat303/cashlog/internal/delivery/http/v1/dto"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"go.uber.org/zap"
)

type transactionUsecase struct {
	txRepo    domain.TransactionRepository
	cacheRepo domain.CacheRepository
	ocrGate   domain.OCRGateway
	log       logger.Logger
}

// Delete implements [domain.TransactionUsecase].
func (t *transactionUsecase) DeleteTransaction(ctx context.Context, id uint) error {
	// ดึง Logger จาก Context ถ้าไม่ม
	log := logger.Ctx(ctx)
	if log == nil {
		// ให้หันไปใช้ Logger ตัวที่สืบทอดมาจาก Constructor แทน ป้องกันแอปพัง (Panic)
		log = t.log
	}

	// ค้นหาข้อมูลเก่า เพื่อส่องดูชื่อไฟล์รูปภาพและช่วงเวลาวันที่
	tx, err := t.txRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// สั่งลบข้อมูลออกจากฐานข้อมูล
	if err := t.txRepo.Delete(ctx, id); err != nil {
		return err
	}

	// ลบ cache ข้อมูลรายเดือนและรายปีออก
	periodKey := fmt.Sprintf("summary:monthly:%d-%02d", tx.TransactionDate.Year(), tx.TransactionDate.Month())
	if err := t.cacheRepo.InvalidateCache(ctx, periodKey); err != nil {
		log.Warn("failed to delete cache from redis ",
			zap.Error(err),
			zap.String("cache_key", periodKey),
		)
	}
	periodYearKey := fmt.Sprintf("summary:yearly:%d", tx.TransactionDate.Year())
	if err := t.cacheRepo.InvalidateCache(ctx, periodYearKey); err != nil {
		log.Warn("failed to delete cache from redis ",
			zap.Error(err),
			zap.String("cache_key", periodYearKey),
		)
	}

	return nil
}

// UpdateTransaction implements [domain.TransactionUsecase].
func (t *transactionUsecase) UpdateTransaction(ctx context.Context, id uint, input dto.UpdateTransactionInput) (*domain.Transaction, error) {
	log := logger.Ctx(ctx)
	if log == nil {
		// ให้หันไปใช้ Logger ตัวที่สืบทอดมาจาก Constructor แทน ป้องกันแอปพัง (Panic)
		log = t.log
	}

	// ตรวจสอบข้อมูลก่อนแก้ไข
	tx, err := t.txRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// บันทึกค่าเวลาเดือนและปีเก่าไว้ก่อนนำไปคำนวณ เพื่อใช้ถล่มแคชกรณีผู้ใช้เปลี่ยนวันที่ข้ามเดือน
	oldYear, oldMonth := tx.TransactionDate.Year(), tx.TransactionDate.Month()

	// เอาช่อมูลชุดใหม่เช้าไป
	if input.Amount != nil {
		tx.Amount = *input.Amount
	}

	if input.Note != nil {
		tx.Note = *input.Note
	}

	if input.CategoryID != nil {
		tx.CategoryID = *input.CategoryID
	}

	if input.TransactionDate != nil {
		tx.TransactionDate = *input.TransactionDate
	}

	// สั่ง update
	if err := t.txRepo.Update(ctx, tx); err != nil {
		return nil, err
	}

	// ทุบแคชของเดือนเก่าและปีเก่าที่เคยบันทึกไว้
	oldPeriodKey := fmt.Sprintf("summary:monthly:%d-%02d", oldYear, oldMonth)
	if err := t.cacheRepo.InvalidateCache(ctx, oldPeriodKey); err != nil {
		log.Warn("failed to delete cache from redis ",
			zap.Error(err),
			zap.String("cache_key", oldPeriodKey),
		)
	}

	// ทุบแคชของเดือนใหม่และปีใหม่ (กรณีผู้ใช้กดเปลี่ยนวันที่ข้ามเดือนย้อนหลัง)
	newPeriodKey := fmt.Sprintf("summary:monthly:%d-%02d", tx.TransactionDate.Year(), tx.TransactionDate.Month())
	if err := t.cacheRepo.InvalidateCache(ctx, newPeriodKey); err != nil {
		log.Warn("failed to delete cache from redis ",
			zap.Error(err),
			zap.String("cache_key", newPeriodKey),
		)
	}

	// ทุบแคชรายปีเพื่อความปลอดภัย
	yearPeriodKey := fmt.Sprintf("summary:yearly:%d", tx.TransactionDate.Year())
	if err := t.cacheRepo.InvalidateCache(ctx, yearPeriodKey); err != nil {
		log.Warn("failed to delete cache from redis ",
			zap.Error(err),
			zap.String("cache_key", yearPeriodKey),
		)
	}

	//  กรณีคนย้ายข้ามปี ก็ต้องทุบแคชของปีเก่าทิ้งด้วยเช่นกัน
	oldYearPeriodKey := fmt.Sprintf("summary:yearly:%d", oldYear)
	if err := t.cacheRepo.InvalidateCache(ctx, oldYearPeriodKey); err != nil {
		log.Warn("failed to delete cache from redis ",
			zap.Error(err),
			zap.String("cache_key", oldYearPeriodKey),
		)
	}

	return tx, nil

}

// FetchByTimeRange implements [domain.TransactionUsecase].
func (t *transactionUsecase) GetMonthlyHistory(ctx context.Context, month, year int) ([]domain.Transaction, error) {

	// กำหนดวันเริ่มต้นของเดือน
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, timeutil.BangKokLoc)

	// กำหนดวันสุดท้ายของเดือน (บวกไป 1 เดือนแล้วหักออก 1 นาโนวินาที)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	// ส่งข่อมูลช่วงเวลาไปให้ Repository
	txs, err := t.txRepo.FetchByTimeRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch monthly trasaction history for %d-%02d: %w", year, month, err)
	}

	return txs, nil
}

// GetDashboardSummary implements [domain.TransactionUsecase].
func (t *transactionUsecase) GetDashboardSummary(ctx context.Context, scope string, month int, year int) (*domain.DashboardSummary, error) {
	log := logger.Ctx(ctx)
	if log == nil {
		// ให้หันไปใช้ Logger ตัวที่สืบทอดมาจาก Constructor แทน ป้องกันแอปพัง (Panic)
		log = t.log
	}
	var startDate, endDate time.Time
	var periodKey string

	if scope == "yearly" {
		// เรื่มวันที่ 1 jan - 31 Dec
		startDate = time.Date(year, time.January, 1, 0, 0, 0, 0, timeutil.BangKokLoc)
		endDate = startDate.AddDate(1, 0, 0).Add(-time.Nanosecond)
		periodKey = fmt.Sprintf("summary:year:%d", year) // ผลลัพธ์ เช่น summary:yearly:2026
	} else {
		// เริ่มวันที่ 1 ของทุกเดือน
		scope = "monthly"
		startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, timeutil.BangKokLoc)
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		periodKey = fmt.Sprintf("summary:monthly:%d-%02d", year, month) // ผลลัพธ์ เช่น summary:monthly:2026-06
	}

	// หาข้อมูล cashe เก่าจาก Redis
	cashedData, err := t.cacheRepo.GetCache(ctx, periodKey)
	if err == nil && cashedData != "" {
		log.Debug("redis cache hit", zap.String("cache_key", periodKey))

		var summary domain.DashboardSummary
		// แปลงข้อความ JSON string จาก Redis กลับมาเป็น Object ่โครงสร้างที่เราต้องการ
		if err := json.Unmarshal([]byte(cashedData), &summary); err == nil {
			return &summary, nil
		}
	}

	log.Debug("redis cache miss, fetching from database",
		zap.String("cache_key", periodKey),
	)

	// หาข้อมูลบน Database
	summary, err := t.txRepo.CalculateSummary(ctx, startDate, endDate, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate summary for key %s: %w", periodKey, err)
	}

	// แปลงข้อมูลที่ได้มาให้เป็น JSon String เพื่อไปเขียนลงใน Redis
	jsonData, err := json.Marshal(summary)
	if err != nil {
		log.Error("failed to marshal dashboard summary for cache",
			zap.Error(err),
			zap.String("cache_key", periodKey),
		)
	} else {
		// ถ้าแปลงเป็น JSON string ได้ เราจะนำไปเขียนใน Redis
		if cacheErr := t.cacheRepo.SetCache(ctx, periodKey, string(jsonData)); cacheErr != nil {
			log.Warn("failed to write dashboard summary to redis cache",
				zap.Error(err),
				zap.String("cache_key", periodKey),
			)
		} else {
			log.Debug("successfully wrote dashboard summary to redis cache",
				zap.String("cache_key", periodKey),
			)
		}
	}

	return summary, nil
}

// SyncTransaction implements [domain.TransactionUsecase].
func (t *transactionUsecase) SyncTransaction(ctx context.Context, imageBytes []byte, localImageName string) (*domain.Transaction, error) {
	log := logger.Ctx(ctx)
	if log == nil {
		// ให้หันไปใช้ Logger ตัวที่สืบทอดมาจาก Constructor แทน ป้องกันแอปพัง (Panic)
		log = t.log
	}
	// ตรวจชื่อไฟล์สลิปบน Redis cache
	isImgProsessed, err := t.cacheRepo.CheckFileExists(ctx, localImageName)
	if err == nil && isImgProsessed {
		// ถ้าเคยแสกนแล้ว เราจะข้ามขั้นตอนไปเลย
		log.Debug("syc transaction early short-circuit (file already processed)",
			zap.String("local_image_name", localImageName),
		)
		return nil, nil
	}

	// ส่งรูปไปให้ AI OCR ประมวลผล
	ocrResult, err := t.ocrGate.Extract(ctx, imageBytes)
	if err != nil {
		return nil, fmt.Errorf("ocr extraction failed: %w", err)
	}

	// จัดการข้อมูลเบื่องต้น
	amount := ocrResult.Amount
	if amount < 0 {
		log.Warn("ocr result constains negative amount, fallback to 0.00",
			zap.String("local_image_name", localImageName),
			zap.Float64("ocr_amount", amount),
		)
		amount = 0.00 // ถ้าอ่านต่าไม่ได้ หรือค่าติดลบกำหนดให้เป็นค่า 0 value
	}

	txDate := ocrResult.TransactionDate
	if txDate.IsZero() {
		txDate = time.Now()
	}

	// นำ TransactionID ไปเช็คความซ่้ำซ้อนใน Database
	if ocrResult.TransactionID != "" {
		isTxDuplicate, err := t.txRepo.CheckDuplicate(ctx, ocrResult.TransactionID)
		if err == nil && isTxDuplicate {
			log.Info("sync transaction skipped (duplicate transaction_id detected)",
				zap.String("local_image_name", localImageName),
				zap.String("transaction_id", ocrResult.TransactionID),
			)

			// ถ้ารหัสซ่้ำ(แตชื่อไฟล์ใหม่) เราจะบันทึกชื่อไฟล์ไว้ใน Redis ไว้ เพื่อป้องกันการแสกนซ้ำรอบหน้า
			if cacheErr := t.cacheRepo.SetFileCache(ctx, localImageName); cacheErr != nil {
				log.Warn("failed to set duplicate file cache in redis",
					zap.Error(cacheErr),
					zap.String("local_image_name", localImageName),
				)
			}
			return nil, nil
		}
	}

	// บันทึก Transaction ใหม่ลงในฐานข้อมูล
	txType := "EXPENSE"
	newTx := &domain.Transaction{
		TransactionID:   ocrResult.TransactionID,
		Amount:          amount,
		TransactionType: txType,
		ReceiverName:    ocrResult.ReceiverName,
		LocalImageName:  localImageName,
		TransactionDate: txDate,
	}

	if err := t.txRepo.Insert(ctx, newTx); err != nil {
		return nil, fmt.Errorf("failed to save trasaction to database: %w", err)
	}

	// เพิ่มข่อมูลในระบบ Cache หลังบันทึกข้อมูลเสร็จ
	if cacheErr := t.cacheRepo.SetFileCache(ctx, localImageName); cacheErr != nil {
		log.Warn("failed to set success file cache in redis",
			zap.Error(err),
			zap.String("local_image_name", localImageName),
		)
	}

	// ล้างแคชแดชบอร์ดสรุปผลประจำเดือน เพื่อตำนวนยอดใหม่
	periodKey := fmt.Sprintf("summary:monthly:%d-%02d", newTx.TransactionDate.Year(), newTx.TransactionDate.Month())
	if invalidateErr := t.cacheRepo.InvalidateCache(ctx, periodKey); invalidateErr != nil {
		log.Warn("failed to invalidate dashboard summary cache",
			zap.Error(invalidateErr),
			zap.String("cache_key", periodKey),
		)
	} else {
		log.Debug("dashboard summary cache invalidate successfully due to new transaction",
			zap.String("cache_key", periodKey),
		)
	}

	return newTx, nil
}

func NewTransactionUsecase(txRepo domain.TransactionRepository, cacheRepo domain.CacheRepository, ocrGateway domain.OCRGateway, log logger.Logger) domain.TransactionUsecase {
	return &transactionUsecase{
		txRepo:    txRepo,
		cacheRepo: cacheRepo,
		ocrGate:   ocrGateway,
		log:       log,
	}
}
