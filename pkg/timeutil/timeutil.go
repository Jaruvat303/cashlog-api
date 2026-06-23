package timeutil

import "time"

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
