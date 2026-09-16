package postgres_test

import (
	"context"
	"testing"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/internal/repository/postgres"
	"github.com/Jaruvat303/cashlog/pkg/database"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSeedCategories_RenameIsIdempotent จำลองฐานข้อมูลจริงที่ยังมีชื่อ category เก่า (ก่อน rename)
// อยู่ แล้วรัน database.SeedCategories — ฟังก์ชันเดียวกับที่ cmd/api/main.go เรียกตอน startup จริง —
// ซ้ำหลายรอบติดกัน เพื่อยืนยันว่า:
//  1. แถวชื่อเก่าถูก rename ในที่เดิม (คง id เดิมไว้) ไม่ใช่ insert แถวใหม่ซ้อนแถวเก่า
//  2. รัน seed ซ้ำกี่รอบก็ตาม ไม่มีชื่อซ้ำ และจำนวนแถวไม่เพิ่มขึ้นหลังรอบแรก
func TestSeedCategories_RenameIsIdempotent(t *testing.T) {
	db := postgres.NewTestDB(t)
	ctx := context.Background()
	log := logger.NewNopLogger()

	// จำลองฐานข้อมูลจริงที่ยังมีชื่อ category เดิม (ก่อน rename ทั้ง 7 รายการ) อยู่ก่อนรัน migration นี้
	// key = ชื่อเก่า, value = ชื่อใหม่ที่ init_categories.sql ต้อง rename ไปให้ถูกต้อง
	renames := map[string]string{
		"เงินบริการและทำบุญ":                 "เงินบริจาคและทำบุญ",
		"ค่าของชำและวัตถุดิบเข้าบ้าน":        "วัตถุดิบเข้าบ้าน",
		"ค่าเดินทางและขนส่งสาธารณะ":          "เดินทางและขนส่งสาธารณะ",
		"ค่าน้ำมันและดูแลรักษารถ":            "น้ำมันและดูแลรักษารถ",
		"ค่าที่อยู่อาศัย (ค่าเช่า/ผ่อนบ้าน)": "ที่อยู่อาศัย (ค่าเช่า/ผ่อนบ้าน)",
		"ค่าอินเทอร์เน็ตและโทรศัพท์":         "อินเทอร์เน็ตและโทรศัพท์",
		"สังสรรค์และปาร์ตี้":                 "ปาร์ตี้และสังสรรค์",
	}

	oldNames := make([]string, 0, len(renames))
	oldIDs := make(map[string]int64, len(renames))
	for oldName := range renames {
		cat := &domain.Category{Name: oldName, Type: domain.TransactionTypeExpense, IconKey: "old-icon", ColorHex: "#000000"}
		require.NoError(t, db.Create(cat).Error)
		oldNames = append(oldNames, oldName)
		oldIDs[oldName] = cat.ID
	}

	// รอบที่ 1
	require.NoError(t, database.SeedCategories(ctx, db, log))

	var countAfterFirst int64
	require.NoError(t, db.Model(&domain.Category{}).Count(&countAfterFirst).Error)
	assert.EqualValues(t, 43, countAfterFirst, "ต้องมี 43 แถวพอดีหลัง seed รอบแรก ไม่ใช่ 43+7 (ถ้า rename กลายเป็น insert ซ้ำ)")

	// รอบที่ 2 และ 3 ติดกัน — ต้อง idempotent เหมือนเดิมทุกประการ
	require.NoError(t, database.SeedCategories(ctx, db, log))
	require.NoError(t, database.SeedCategories(ctx, db, log))

	var countAfterRepeat int64
	require.NoError(t, db.Model(&domain.Category{}).Count(&countAfterRepeat).Error)
	assert.Equal(t, countAfterFirst, countAfterRepeat, "รัน seed ซ้ำหลายรอบติดกันต้องไม่สร้างแถวเพิ่ม")

	// ชื่อเก่าต้องไม่มีเหลืออยู่ในตารางแล้ว (ถูก rename ไปหมด)
	var oldNameCount int64
	require.NoError(t, db.Model(&domain.Category{}).Where("name IN ?", oldNames).Count(&oldNameCount).Error)
	assert.Zero(t, oldNameCount, "ต้องไม่มีชื่อเก่าเหลืออยู่ในตารางหลัง seed")

	// ต้องไม่มีชื่อ category ซ้ำกันเลยในตารางทั้งหมด
	var names []string
	require.NoError(t, db.Model(&domain.Category{}).Pluck("name", &names).Error)
	seen := make(map[string]bool, len(names))
	for _, n := range names {
		assert.False(t, seen[n], "พบชื่อ category ซ้ำ: %s", n)
		seen[n] = true
	}

	// id เดิมของแถวที่ rename ต้องยังเป็น id เดิมสำหรับทั้ง 7 รายการ (พิสูจน์ว่าเป็นการ UPDATE ในที่เดิม ไม่ใช่ insert แถวใหม่)
	for oldName, newName := range renames {
		var renamed domain.Category
		require.NoError(t, db.Where("name = ?", newName).First(&renamed).Error, "ต้องเจอแถวชื่อใหม่ %q หลัง seed", newName)
		assert.Equal(t, oldIDs[oldName], renamed.ID, "id ของ %q -> %q ต้องคงเดิมหลัง rename ไม่ใช่แถวใหม่", oldName, newName)
	}
}
