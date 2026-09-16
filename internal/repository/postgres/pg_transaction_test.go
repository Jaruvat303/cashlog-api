package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/internal/repository/postgres"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGormTransactionRepository_Update_ClearsCategoryID จำลองสิ่งที่ Usecase ticket 04 ทำตอนแปลง
// transaction_type: เคลียร์ CategoryID เป็น nil บน struct ที่ preload Category ติดมาจาก GetByID แล้วเรียก Update
//
// นี่คือเทสต์กัน regression ของบั๊กที่เจอระหว่าง manual verify ticket 04: GORM Save() เคย sync category_id
// กลับจาก tx.Category (association struct ที่ preload ค้างมาจาก GetByID) ทับค่า nil ที่เพิ่งเคลียร์ไป —
// ต้องแดงถ้ามีใครเอา .Omit(clause.Associations) ออกจาก gormTransactionRepository.Update() ในอนาคต
func TestGormTransactionRepository_Update_ClearsCategoryID(t *testing.T) {
	db := postgres.NewTestDB(t)
	ctx := context.Background()
	log := logger.NewNopLogger()

	// seed: account 1 ตัว, category 2 ตัว (คนละ Type กัน)
	account := &domain.Account{Name: "บัญชีทดสอบ", AccountType: "bank", IsActive: true}
	require.NoError(t, db.Create(account).Error)

	expenseCategory := &domain.Category{Name: "หมวดรายจ่ายทดสอบ", Type: domain.TransactionTypeExpense}
	require.NoError(t, db.Create(expenseCategory).Error)

	incomeCategory := &domain.Category{Name: "หมวดรายรับทดสอบ", Type: domain.TransactionTypeIncome}
	require.NoError(t, db.Create(incomeCategory).Error)

	// seed: transaction type=expense ผูก category ฝั่ง expense อยู่
	seedTx := &domain.Transaction{
		Amount:          100,
		TransactionType: domain.TransactionTypeExpense,
		AccountID:       &account.ID,
		CategoryID:      &expenseCategory.ID,
		TransactionDate: time.Now(),
	}
	require.NoError(t, db.Create(seedTx).Error)

	repo := postgres.NewGormTransactionRepository(db, log)

	// ดึง transaction ด้วย GetByID ให้ preload Category ติดมาเหมือนที่ usecase ทำจริง
	fetched, err := repo.GetByID(ctx, seedTx.ID)
	require.NoError(t, err)
	assert.Equal(t, expenseCategory.ID, fetched.Category.ID, "ต้อง preload Category ติดมาด้วยก่อนเริ่มเทสต์")

	// จำลองสิ่งที่ usecase ticket 04 ทำตอนแปลง type: เซ็ต CategoryID เป็น nil บน struct ที่ได้
	// (tx.Category ยังเป็น struct เดิมที่ preload ค้างอยู่ — ไม่ได้ถูกเคลียร์ ตรงนี้คือจุดที่บั๊กเคยเกิด)
	fetched.CategoryID = nil

	err = repo.Update(ctx, fetched)
	require.NoError(t, err)

	// อ่านกลับจาก DB ตรงๆ อีกครั้งด้วย instance คนละตัว เพื่อไม่ให้ปนกับตัวแรก
	reloaded, err := repo.GetByID(ctx, seedTx.ID)
	require.NoError(t, err)

	assert.Nil(t, reloaded.CategoryID, "category_id ต้องถูกเคลียร์เป็น nil จริงใน DB หลัง Update")
	assert.Zero(t, reloaded.Category.ID, "ไม่ควรมี Category ติดมาด้วยหลัง category_id ถูกเคลียร์แล้ว")
}
