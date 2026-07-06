package timeutil

import (
	"time"
)

// ประกาศตัวแปร Global สำหรับเก็บเวลา Asia/Bangkok
var BangKokLoc *time.Location

// กลไกของ Go จะรันฟังก์ชัน init() นี้ให้อัตโนมัติทันทีที่มีเลเยอร์ไหนอ้างอิงถึงแพ็คเกจนี้
func init() {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		BangKokLoc = time.Local
	}
	BangKokLoc = loc
}

// NowInBangkok แปลงเวลาที่ได้มาเป็นเวลา Asia/Bangkok
func NowInBangkok() time.Time {
	return time.Now().In(BangKokLoc)
}

// ในส่วนของ repository หรือ usecase หลังจากได้ result จาก Gemini
func ParseAISlipTime(aiTimeString string) (time.Time, error) {
	// ใช้รูปแบบที่ตกลงกับ AI คือ "2006-01-02 15:04:05"
	// บอก Go ว่านี่คือเวลาใน Timezone Bangkok ตั้งแต่ต้น
	return time.ParseInLocation("2006-01-02 15:04:05", aiTimeString, BangKokLoc)
}
