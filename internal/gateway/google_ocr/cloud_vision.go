package googleocr

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	vision "cloud.google.com/go/vision/v2/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"github.com/Jaruvat303/cashlog/internal/domain"
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
)

type googleOCRGateway struct {
}

// Extract implements [domain.OCRGateway].
func (g *googleOCRGateway) Extract(ctx context.Context, imageBytes []byte) (*domain.OCRData, error) {
	// เปิดการเชื่อมต่อ Client กับ Google Cloud Vision API
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create vision client: %w", err)
	}
	defer client.Close()

	// ใช้โตรงสร้าง visionpb.Image และใส่ content เป็น byte เป็น byte array โดยตรง
	image := &visionpb.Image{
		Content: imageBytes,
	}

	// สร้าง Request เพื่อระบุว่าต้องการทำ TEXT_DETECTION (OCR)
	imageReq := &visionpb.AnnotateImageRequest{
		Image: image,
		Features: []*visionpb.Feature{
			{
				Type: visionpb.Feature_TEXT_DETECTION, // สั่งให้ทำการตรวจจับและสกัดข้อความออกจากรูปภ่าพ
			},
		},
	}

	// ครอบ Request ย่อยลงใน BatchAnnotateImagesRequest
	req := &visionpb.BatchAnnotateImagesRequest{
		Requests: []*visionpb.AnnotateImageRequest{imageReq}, // ส่งอาเรย์ที่มีรูปเดียวเข้าไป
	}

	// เรียกใช้งานฟังก์ชัน BatchAnnotateImages เพื่อประมวลผลรูปภาพ
	resp, err := client.BatchAnnotateImages(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to batch annotate image via google vision api: %w", err)
	}

	// ตรวจสอบและสกัดดึงข้อมูลผลลัพธ์ออกมาจากโครงสร้าง Batch Response
	response := resp.GetResponses()
	if len(response) == 0 {
		return &domain.OCRData{}, nil
	}

	// ดึงผลลัพท์ของรูปแรกที่เราส่งไป
	firstResp := response[0]
	if firstResp.GetError() != nil {
		return nil, fmt.Errorf("google vision api internal error: %s", firstResp.GetError().GetMessage())
	}

	annotations := firstResp.GetTextAnnotations()
	// หาก AI ตรวจไม่พบข้อมูลข้อความใดๆ และส่งค่าว่างกลับไป
	if len(annotations) == 0 {
		return &domain.OCRData{}, nil
	}

	// ข้อความดิบทั้งหมดที่ AI อ่านได้จากสลืปจะรวมอยู่ใน Description ของดัชนีแรก [0]
	fullText := annotations[0].Description

	// 🚨 [เพิ่มตรงนี้] พิมพ์ข้อความดิบออกทางหน้าจอ เพื่อดูว่า AI อ่านสลิปใบนี้ออกมาเป็นยังไง!
	fmt.Println("========== 🤖 VISION AI RAW TEXT ==========")
	fmt.Println(fullText)
	fmt.Println("===========================================")

	// นำข้อความดิบไปเข้ากับบวนการสกัดข้อมูล (Data Parsing)
	ocrData := g.parseSlipText(fullText)

	return ocrData, nil

}

// parseSlipText ตรรกะคัดกรองและดึงข้อมูลด้วย Regex ที่ปรับแต่งมาเพื่อสลิปธนาคาร SCB โดยเฉพาะ
func (g *googleOCRGateway) parseSlipText(text string) *domain.OCRData {
	data := &domain.OCRData{
		TransactionDate: time.Now(), // ค่าเริ่มต้นเผื่อกรณีแกะวันเวลาไม่สำเร็จ
	}

	// 1. ล้างช่องว่างที่อาจเกิดขึ้นแปลกๆ จาก OCR และแยกเป็นทีละบรรทัด
	lines := strings.Split(text, "\n")

	// 🛠️ 2. กำหนดโครงสร้าง Regex ที่เจาะจงพฤติกรรมสลิป SCB

	// เครื่องสแกนรูปร่าง: วันเวลา
	// ดักจับ: "04 ก.ค. 2569 - 10:15" หรือ "4 ก.ค. 69 - 10:15"
	dateRegex := regexp.MustCompile(`([0-9]{1,2})\s*([ก-์A-Za-z\.]+)\s*([0-9]{2,4})\s*-\s*([0-9]{2}:[0-9]{2})`)

	// 2. Regex ดักจับรูปแบบตัวเลขจำนวนเงินที่อยู่เดี่ยวๆ ใน 1 บรรทัด เช่น "57.00" หรือ "1,250.00"
	// รูปแบบ A: ตัวเลขโดดๆ ที่มีทศนิยม .00 (เช่น "57.00" ซึ่งมักจะอยู่บรรทัดล่างสุด)
	amountRegex := regexp.MustCompile(`^([0-9]{1,3}(?:,[0-9]{3})*\.[0-9]{2})$`)	
	
	// ตัวแปรจำสถานะบรรทัด เพื่อช่วยแกะชื่อผู้รับเงิน
	isNextLineReceiver := false
	isNextLineAmount := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// ดึงข้อมูล วัน-เวลา และแปลงเป็น time.Time (พยายามแกะโครงสร้างเวลาสลิป)
		if dateMatches := dateRegex.FindStringSubmatch(line); len(dateMatches) > 4 {
			dayStr, monthStr, yearStr, timeStr := dateMatches[1], dateMatches[2], dateMatches[3], dateMatches[4]

			var day, year, hour, min int
			fmt.Sscanf(dayStr, "%d", &day)
			fmt.Sscanf(yearStr, "%d", &year)
			fmt.Sscanf(timeStr, "%d:%d", &hour, &min)

			month, err := timeutil.ParseThaiMonthAbbr(monthStr)
			if err != nil {
				month = timeutil.NowInBangkok().Month()
			}

			if year > 2500 {
				year -= 543 // แปลง พ.ศ. เป็น ค.ศ.
			}

			data.TransactionDate = time.Date(year, month, day, hour, min, 0, 0, timeutil.BangKokLoc)
		}

		// ดึงจำนวนเงิน (Amount)
		if data.Amount == 0 {
			// กรณีบรรทัดนี้คือคำว่า "จำนวนเงิน" หรือตำว่า "ยอดเงิน"
			if strings.Contains(line, "จำนวนเงิน") || strings.Contains(line, "ยอดเงิน") {
				// หลังคำว่า "จำนวนเงิน" หรือตำว่า "ยอดเงิน" ลองหาตัวเลขในบรรทัดเดียวกันก่อน
				if amtMatch := amountRegex.FindStringSubmatch(line); len(amtMatch) > 1 {
					data.Amount = parseAmountString(amtMatch[1])
				} else {
					isNextLineAmount = true
				}
				continue
			}

			// กรณีเป็นบรรทัดที่ตามหลังตำว่า "จำนวนเงิน
			if isNextLineAmount {
				if amtMatch := amountRegex.FindStringSubmatch(line); len(amtMatch) > 1 {
					data.Amount = parseAmountString(amtMatch[1])
					isNextLineAmount = false
					continue
				}
			}
		}

		// ดึงชื่อผู้รับเงิน (Receiver Name)
		if data.ReceiverName == "" {
			if strings.Contains(line, "ไปยัง") {
				// ลบคำว่า "ไปยัง" ออก เผื่อ OCR อ่านติดมาในบรรทัดเดียวกัน
				namePart := strings.TrimSpace(strings.Replace(line, "ไปยัง", "", 1))

				if namePart != "" {
					data.ReceiverName = cleanReceiverName(namePart)
				} else {
					// ถ้าคำว่า ไปยัง อยู่เดี่ยวๆ ให้รออ่านชื่อในบรรทัดถัดไป
					isNextLineReceiver = true
				}

				continue
			}

			if isNextLineReceiver {
				data.ReceiverName = cleanReceiverName(line)
				isNextLineReceiver = false
				continue
			}
		}

	}

	return data
}

// NewGoogleOCRGateway ทำหน้าที่สร้างอินสแตนซ์สำหรับเรียกใช้บริการ OCR
func NewGoogleOCRGateway() domain.OCRGateway {
	return &googleOCRGateway{}
}

// 🛠️ ฟังก์ชันตัวช่วย: แปลงข้อความจำนวนเงินเป็น float64
func parseAmountString(amtStr string) float64 {
	cleanStr := strings.ReplaceAll(amtStr, ",", "")
	val, _ := strconv.ParseFloat(cleanStr, 64)
	return val
}

// 🛠️ ฟังก์ชันตัวช่วย: ตัดคำว่า "SCB มณี SHOP (...)" ออกเพื่อเอาแค่ชื่อร้านข้างใน
func cleanReceiverName(name string) string {
	if strings.Contains(name, "SCB มณี SHOP") {
		// ใช้ Regex ดึงข้อความที่อยู่ในวงเล็บ
		re := regexp.MustCompile(`\((.*?)\)`)
		if matches := re.FindStringSubmatch(name); len(matches) > 1 {
			return strings.TrimSpace(matches[1]) // จะได้คำว่า "เฟชรมาร์ท" ออกมาเพียวๆ
		}
	}
	return name
}
