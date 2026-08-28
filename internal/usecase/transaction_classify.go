package usecase

import (
	"strings"

	"github.com/Jaruvat303/cashlog/internal/domain"
)

// matchesOwnerAlias เช็คว่าชื่อจากสลิปตรงกับเจ้าของบัญชี (OWNER_NAME_ALIASES) หรือไม่
// เทียบแบบไม่สนตัวพิมพ์เล็ก-ใหญ่ และยอมรับ partial match (ชื่อจากสลิปมี alias เป็น substring)
func matchesOwnerAlias(name string, ownerAliases []string) bool {
	if name == "" {
		return false
	}

	lowerName := strings.ToLower(name)
	for _, alias := range ownerAliases {
		if alias == "" {
			continue
		}
		if strings.Contains(lowerName, strings.ToLower(alias)) {
			return true
		}
	}

	return false
}

// classifyTransactionType จำแนกประเภทธุรกรรมจากชื่อผู้ส่ง/ผู้รับบนสลิปเทียบกับ OWNER_NAME_ALIASES (BR-1)
//
// 4 branch ตาม Decision #26:
//  1. sender ตรงเจ้าของ + receiver ไม่ตรง -> expense
//  2. receiver ตรงเจ้าของ + sender ไม่ตรง -> income
//  3. ทั้งคู่ตรงเจ้าของ -> transfer
//  4. ไม่ตรงทั้งคู่ -> expense (fallback)
func classifyTransactionType(senderName, receiverName string, ownerAliases []string) string {
	senderIsOwner := matchesOwnerAlias(senderName, ownerAliases)
	receiverIsOwner := matchesOwnerAlias(receiverName, ownerAliases)

	switch {
	case senderIsOwner && receiverIsOwner:
		return domain.TransactionTypeTransfer
	case receiverIsOwner && !senderIsOwner:
		return domain.TransactionTypeIncome
	default:
		// ครอบคลุมทั้ง branch 1 (sender=owner, receiver!=owner) และ branch 4 (ไม่ตรงทั้งคู่)
		return domain.TransactionTypeExpense
	}
}

// matchAccountByKeyword จับคู่ app_name จากสลิปกับ MatchingKeywords ของบัญชีที่ active อยู่ (BR-2)
//
// เทียบแบบไม่สนตัวพิมพ์เล็ก-ใหญ่ + partial match (app_name มี keyword เป็น substring)
// ถ้ามีหลาย keyword match พร้อมกัน เลือก keyword ที่ยาว/เจาะจงที่สุด (Decision #5)
// ไม่มี match คืนค่า nil (nullable, ปล่อยให้ user PATCH เอง ตาม Decision #18)
func matchAccountByKeyword(appName string, activeAccounts []domain.Account) *int64 {
	if appName == "" {
		return nil
	}

	lowerAppName := strings.ToLower(appName)

	var matchedID *int64
	bestKeywordLen := 0

	for _, acc := range activeAccounts {
		for _, keyword := range acc.MatchingKeywords {
			if keyword == "" {
				continue
			}
			if !strings.Contains(lowerAppName, strings.ToLower(keyword)) {
				continue
			}
			if len(keyword) > bestKeywordLen {
				bestKeywordLen = len(keyword)
				id := acc.ID
				matchedID = &id
			}
		}
	}

	return matchedID
}
