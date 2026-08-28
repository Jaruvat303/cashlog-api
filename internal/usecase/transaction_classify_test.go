package usecase

import (
	"testing"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/stretchr/testify/assert"
)

// TestClassifyTransactionType ทดสอบ BR-1 (auto-classify) ครบ 4 branch ตาม Decision #26
func TestClassifyTransactionType(t *testing.T) {
	ownerAliases := []string{"jaruvat", "จารุวัฒน์"}

	tests := []struct {
		name         string
		senderName   string
		receiverName string
		ownerAliases []string
		expected     string
	}{
		{
			name:         "1. Branch 1: sender ตรงเจ้าของ + receiver ไม่ตรง -> expense",
			senderName:   "Jaruvat Seesuwan",
			receiverName: "ร้านข้าวมันไก่ป้าใจ",
			ownerAliases: ownerAliases,
			expected:     domain.TransactionTypeExpense,
		},
		{
			name:         "2. Branch 2: receiver ตรงเจ้าของ + sender ไม่ตรง -> income",
			senderName:   "ร้านกาแฟ",
			receiverName: "นายจารุวัฒน์ ส.",
			ownerAliases: ownerAliases,
			expected:     domain.TransactionTypeIncome,
		},
		{
			name:         "3. Branch 3: ทั้งคู่ตรงเจ้าของ -> transfer",
			senderName:   "Jaruvat S.",
			receiverName: "จารุวัฒน์ ส.",
			ownerAliases: ownerAliases,
			expected:     domain.TransactionTypeTransfer,
		},
		{
			name:         "4. Branch 4: ไม่ตรงทั้งคู่ -> expense (fallback)",
			senderName:   "ร้านค้า A",
			receiverName: "ร้านค้า B",
			ownerAliases: ownerAliases,
			expected:     domain.TransactionTypeExpense,
		},
		{
			name:         "5. Case-insensitive: sender ตัวพิมพ์ใหญ่ล้วนยังต้อง match",
			senderName:   "JARUVAT SEESUWAN",
			receiverName: "ร้านค้า B",
			ownerAliases: ownerAliases,
			expected:     domain.TransactionTypeExpense,
		},
		{
			name:         "6. Partial match: alias เป็น substring ของชื่อเต็มบนสลิป",
			senderName:   "Mr. Jaruvat Seesuwan (Silver)",
			receiverName: "ร้านค้า B",
			ownerAliases: ownerAliases,
			expected:     domain.TransactionTypeExpense,
		},
		{
			name:         "7. ชื่อว่างทั้งคู่ -> expense (fallback, ไม่ crash)",
			senderName:   "",
			receiverName: "",
			ownerAliases: ownerAliases,
			expected:     domain.TransactionTypeExpense,
		},
		{
			name:         "8. ไม่มี OWNER_NAME_ALIASES เลย -> expense (fallback) แม้ชื่อจะดูเหมือนเจ้าของ",
			senderName:   "Jaruvat Seesuwan",
			receiverName: "ร้านค้า B",
			ownerAliases: []string{},
			expected:     domain.TransactionTypeExpense,
		},
		{
			name:         "9. alias ว่างปนอยู่ในลิสต์ ต้องข้ามไปไม่ crash และ match ตัวอื่นได้ปกติ",
			senderName:   "จารุวัฒน์ ส.",
			receiverName: "ร้านค้า B",
			ownerAliases: []string{"", "จารุวัฒน์"},
			expected:     domain.TransactionTypeExpense,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyTransactionType(tt.senderName, tt.receiverName, tt.ownerAliases)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMatchAccountByKeyword ทดสอบ BR-2 (auto-match account) รวม conflict resolution ตาม Decision #5
func TestMatchAccountByKeyword(t *testing.T) {
	ptr := func(id int64) *int64 { return &id }

	activeAccounts := []domain.Account{
		{ID: 1, MatchingKeywords: domain.StringSlice{"SCB", "ไทยพาณิชย์"}},
		{ID: 2, MatchingKeywords: domain.StringSlice{"SCB EASY"}},
		{ID: 3, MatchingKeywords: domain.StringSlice{"", "Dime"}},
	}

	tests := []struct {
		name           string
		appName        string
		activeAccounts []domain.Account
		expected       *int64
	}{
		{
			name:           "1. app_name ว่าง -> ไม่ match (nil)",
			appName:        "",
			activeAccounts: activeAccounts,
			expected:       nil,
		},
		{
			name:           "2. ไม่มี keyword ไหน match เลย -> nil",
			appName:        "K PLUS",
			activeAccounts: activeAccounts,
			expected:       nil,
		},
		{
			name:           "3. match ตรงตัว keyword เดียว -> คืน account ID นั้น",
			appName:        "SCB",
			activeAccounts: activeAccounts,
			expected:       ptr(1),
		},
		{
			name:           "4. Conflict: 2 บัญชี match พร้อมกัน -> เลือก keyword ที่ยาว/เจาะจงที่สุด (Decision #5)",
			appName:        "SCB EASY",
			activeAccounts: activeAccounts,
			expected:       ptr(2), // "SCB EASY" (8 ตัวอักษร) ยาวกว่า "SCB" (3 ตัวอักษร) ของบัญชี 1
		},
		{
			name:           "5. Case-insensitive + partial match",
			appName:        "เปิดแอป scb easy บนมือถือ",
			activeAccounts: activeAccounts,
			expected:       ptr(2),
		},
		{
			name:           "6. keyword ว่างในลิสต์ต้องถูกข้าม ไม่ crash และ match keyword อื่นในบัญชีเดียวกันได้ปกติ",
			appName:        "Dime",
			activeAccounts: activeAccounts,
			expected:       ptr(3),
		},
		{
			name:           "7. ไม่มีบัญชี active เลย -> nil",
			appName:        "SCB",
			activeAccounts: []domain.Account{},
			expected:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchAccountByKeyword(tt.appName, tt.activeAccounts)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				if assert.NotNil(t, result) {
					assert.Equal(t, *tt.expected, *result)
				}
			}
		})
	}
}
