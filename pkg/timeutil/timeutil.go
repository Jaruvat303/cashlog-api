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

// MonthRangeBangkok คืนค่าช่วงเวลาแบบ half-open [start, end) ของเดือนที่ระบุ
// ตามเวลา Asia/Bangkok โดย start คือ 00:00 ของวันที่ 1 และ end คือ 00:00 ของวันที่ 1
// เดือนถัดไป (รองรับการข้ามปี เช่น เดือน 12 -> end เป็นเดือน 1 ปีถัดไป)
//
// month ต้องอยู่ในช่วง 1-12 เท่านั้น ผู้เรียกต้อง validate เอง: time.Date
// จะ normalize เดือนที่อยู่นอกช่วงแบบเงียบๆ (เช่น month=13 กลายเป็นมกราคมปีถัดไป,
// month=0 กลายเป็นธันวาคมปีก่อนหน้า) โดยไม่ error
func MonthRangeBangkok(year int, month int) (start, end time.Time) {
	start = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, BangKokLoc)
	end = start.AddDate(0, 1, 0)
	return start, end
}

// YearRangeBangkok คืนค่าช่วงเวลาแบบ half-open [start, end) ของปีที่ระบุ
// ตามเวลา Asia/Bangkok โดย start คือ 00:00 ของวันที่ 1 มกราคม และ end คือ
// 00:00 ของวันที่ 1 มกราคมปีถัดไป
func YearRangeBangkok(year int) (start, end time.Time) {
	start = time.Date(year, time.January, 1, 0, 0, 0, 0, BangKokLoc)
	end = start.AddDate(1, 0, 0)
	return start, end
}
