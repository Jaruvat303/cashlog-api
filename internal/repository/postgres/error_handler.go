package postgres

import (
	"context"
	"errors"
	"runtime"
	"strings"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HandlerDBError ทำหน้าที่แปลง Error จาก Database Driver ให้เป็น Domain.Error
func HandlerDBError(ctx context.Context, err error, appLogger logger.Logger) error {
	if err == nil {
		return nil
	}

	//  ดึง Logger จาก Context ก่อน
	log := logger.Ctx(ctx)

	//  ถ้าใน Context ไม่มี (เป็น nil) ให้สลับไปใช้ appLogger ที่ส่งมาทันที
	if log == nil {
		log = appLogger
	}

	// ใช้ runtime.Caller(1) เพื่อดึงข้อมูลของฟังก์ชัน "ก่อนหน้า" (คนที่มีเรียกฟังก์ชันนี้)
	fnName := "unknown"
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		details := runtime.FuncForPC(pc)
		if details != nil {
			fullFnName := details.Name()
			parts := strings.Split(fullFnName, "/")
			fnName = parts[len(parts)-1]
		}
	}

	// 1. ตรวจสอบ Context Errors ก่อน (User กดยกเลิก หรือ Timeout จากฝั่ง Client/Gateway)
	if errors.Is(err, context.Canceled) {
		log.Warn("request canceled by user", zap.String("func", fnName))
		return domain.ErrContextCanceled
	}

	if errors.Is(err, context.DeadlineExceeded) {
		// เกิดจาก Database ทำงานช้ากว่า Timeout ที่เราตั้งไว้ใน Context
		log.Error("database query timeout exceeded",
			zap.String("func", fnName),
			zap.Error(err),
		)
		return domain.ErrInternalDB // หรือ domain.ErrTimeout
	}

	//  ตรวจสอบ GORM Error ทั่วไป
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrNotFound
	}

	// ตรวจสอบ Postgres Driver Error (เช่น Unique Constraint, Foreign Key)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // Unique Violation
			return domain.ErrDuplicateRequest
		case "23503":
			log.Warn("foreign key constraint violation",
				zap.String("func", fnName),
				zap.Error(err),
				zap.String("table", pgErr.TableName),
			)
			return domain.ErrInternalDB
		}
	}

	// Error ที่เราไม่คาดติด (เช่น SQL Syntax ผิด, DB ล่ม, Supabase Connection เต็ม)
	log.Error("unexpected database error",
		zap.String("func", fnName),
		zap.Error(err),
	)

	return domain.ErrInternalDB
}
