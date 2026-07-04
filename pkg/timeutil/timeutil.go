package timeutil

import (
	"errors"
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

// ParseThaiMonthAbbr รับค่าตัวย่อเดือนภาษาไทยและแปลงเป็น time.Month
func ParseThaiMonthAbbr(abbr string) (time.Month, error) {
	switch abbr {
	case "ม.ค.":
		return time.January, nil
	case "ก.พ.":
		return time.February, nil
	case "มี.ค.":
		return time.March, nil
	case "เม.ย.":
		return time.April, nil
	case "พ.ค.":
		return time.May, nil
	case "มิ.ย.":
		return time.June, nil
	case "ก.ค.":
		return time.July, nil
	case "ส.ค.":
		return time.August, nil
	case "ก.ย.":
		return time.September, nil
	case "ต.ค.":
		return time.October, nil
	case "พ.ย.":
		return time.November, nil
	case "ธ.ค.":
		return time.December, nil
	default:
		// ส่งค่า 0 และ error กลับไปหากไม่ตรงกับตัวย่อใดเลย
		return 0, errors.New("invalid Thai month abbreviation")
	}
}
