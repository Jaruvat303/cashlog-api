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

	// Transaction ID ของ SCB มักจะเจอคำว่า "เลขที่อ้างอิง" หรือ "รหัสรายการ"
	// และตามด้วยรหัสยาวๆ (เช่น 20260531xxxxxxxx) หรือบางครั้ง Google OCR อ่านสลับบรรทัด จึงดักจับเลขชุดยาว 16-20 หลักไว้ด้วย
	txIDRegex := regexp.MustCompile(`(?:เลขที่อ้างอิง|รหัสรายการ|Ref\.?\s*No\.?)[:\s]*([A-Za-z0-9]{15,20})|([0-9]{16,20})`)

	// จำนวนเงินของ SCB มักจะอยู่บรรทัดเดียวกับคำว่า "จำนวนเงิน" หรือ "Amount"
	amountRegex := regexp.MustCompile(`(?:จำนวนเงิน|ยอดเงิน|Amount)[:\s]*([0-9,]+\.[0-9]{2})|([0-9,]+\.[0-9]{2})\s*(?:บาท|THB)`)

	// วันเวลาในสลิป SCB มักมาในรูปแบบ "31 พ.ค. 2569 - 23:30" หรือ "31 May 2026, 23:30"
	// เราใช้ Regex ช่วยดักจับโครงสร้าง วัน เดือน ปี และ เวลา ออกมา
	dateRegex := regexp.MustCompile(`([0-9]{1,2})\s*([ก-์A-Za-z\.]+)\s*([0-9]{2,4})[\s,-]*([0-9]{2}:[0-9]{2})`)

	// ตัวแปรจำสถานะบรรทัด เพื่อช่วยแกะชื่อผู้รับเงิน
	isNextLineReceiver := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 3.1 ดึงรหัส Transaction ID
		if data.TransactionID == "" {
			if txIDMatches := txIDRegex.FindStringSubmatch(line); len(txIDMatches) > 0 {
				// เลือก Match Group ตัวที่ไม่ว่าง
				if txIDMatches[1] != "" {
					data.TransactionID = txIDMatches[1]
				} else if txIDMatches[2] != "" {
					data.TransactionID = txIDMatches[2]
				}
			}
		}

		// 3.2 ดึงจำนวนเงิน (Amount)
		if data.Amount == 0 {
			if amountMatches := amountRegex.FindStringSubmatch(line); len(amountMatches) > 0 {
				var amtStr string
				if amountMatches[1] != "" {
					amtStr = amountMatches[1]
				} else if amountMatches[2] != "" {
					amtStr = amountMatches[2]
				}

				if amtStr != "" {
					cleanAmountStr := strings.ReplaceAll(amtStr, ",", "")
					if parsedAmount, err := strconv.ParseFloat(cleanAmountStr, 64); err == nil {
						data.Amount = parsedAmount
					}
				}
			}
		}

		// 3.3 ดึงข้อมูล วัน-เวลา และแปลงเป็น time.Time (พยายามแกะโครงสร้างเวลาสลิป)
		if dateMatches := dateRegex.FindStringSubmatch(line); len(dateMatches) > 4 && data.TransactionDate.Year() == time.Now().Year() {
			// แกะข้อมูลเวลาชั่วโมง:นาที มาใช้งานก่อนเพราะแม่นยำสุด
			timeStr := dateMatches[4] // "23:30"
			var hour, min int
			fmt.Sscanf(timeStr, "%d:%d", &hour, &min)

			// นำมาประกอบร่างโครงสร้างเวลาพื้นฐาน (สำหรับความสมบูรณ์ในการบันทึก DB)
			data.TransactionDate = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hour, min, 0, 0, time.Local)
		}

		// 3.4 ดึงชื่อผู้รับเงิน (Receiver Name)
		// สลิป SCB จะมีคำสำคัญบอกว่าเงินโอนไปหาใคร เช่นบรรทัดก่อนหน้าเขียนว่า "เข้าบัญชี" หรือ "ไปยัง"
		// หรือในบรรทัดนั้นมีคำระบุตำแหน่งบุคคล/ร้านค้าปลายทาง
		if strings.Contains(line, "เข้าบัญชี") || strings.Contains(line, "ไปยัง") || strings.Contains(line, "โอนเงินสำเร็จ") {
			isNextLineReceiver = true
			continue
		}

		if isNextLineReceiver && data.ReceiverName == "" {
			// คัดกรองบรรทัดถัดมาที่ไม่ใช่เลขบัญชี หรือคำว่า บาท ให้เป็นชื่อผู้รับเงิน
			if !strings.Contains(line, "xxx-x") && !strings.Contains(line, "บาท") {
				data.ReceiverName = line
				isNextLineReceiver = false // แกะได้แล้วปิดสวิตช์
			}
		}

		// กรณีระบบเดาแบบ Fallback เพิ่มเติม หากบรรทัดข้างบนหลุดรอดไป
		if data.ReceiverName == "" {
			if strings.Contains(line, "นาย ") || strings.Contains(line, "นางสาว ") || strings.Contains(line, "บริษัท ") || strings.Contains(line, "บจก. ") {
				data.ReceiverName = line
			}
		}
	}

	return data
}

// NewGoogleOCRGateway ทำหน้าที่สร้างอินสแตนซ์สำหรับเรียกใช้บริการ OCR
func NewGoogleOCRGateway() domain.OCRGateway {
	return &googleOCRGateway{}
}
