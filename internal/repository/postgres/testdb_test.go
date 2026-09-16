package postgres

import (
	"testing"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewTestDB สร้าง GORM DB ทดสอบต่อกับ SQLite in-memory พร้อม AutoMigrate schema ที่เกี่ยวข้อง
// (domain.Account, domain.Category, domain.Transaction) — ใช้ร่วมกันได้ทั้งเทสต์ white-box
// ในแพ็กเกจ postgres และเทสต์ black-box ในแพ็กเกจ postgres_test (เรียกผ่าน postgres.NewTestDB)
//
// จำกัด connection pool ไว้ที่ 1 เชื่อมต่อเสมอ: driver mattn/go-sqlite3 เปิด ":memory:" ใหม่แยกกันทุกครั้งที่มี
// connection ใหม่ในพูล ถ้าปล่อยให้ gorm เปิดหลาย connection พร้อมกัน แต่ละอันจะเห็นฐานข้อมูลคนละก้อน (schema/ข้อมูลหาย)
func NewTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	err = db.AutoMigrate(&domain.Account{}, &domain.Category{}, &domain.Transaction{})
	require.NoError(t, err)

	return db
}
