package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Jaruvat303/cashlog/internal/domain"
	"google.golang.org/genai"
)

// slipScanPrompt คือ instruction ที่บอก Gemini ว่าต้องอ่านฟิลด์ไหนจากสลิป
// และกฎการตอบเมื่ออ่านไม่ออก (ห้ามเดา/แต่งข้อมูล)
const slipScanPrompt = `คุณคือระบบอ่านข้อมูลจากภาพสลิปโอนเงินผ่าน mobile banking ของธนาคารไทย

จากภาพที่ให้มา ให้ดึงข้อมูลต่อไปนี้:
- amount: จำนวนเงินที่โอน (ตัวเลขล้วน ไม่มีคอมม่าหรือสัญลักษณ์สกุลเงิน)
- sender_name: ชื่อบัญชีต้นทาง (ผู้โอน) ตามที่ปรากฏในสลิป
- receiver_name: ชื่อบัญชีปลายทาง (ผู้รับเงิน) ตามที่ปรากฏในสลิป
- trans_time: วันและเวลาที่ทำรายการ แปลงเป็นรูปแบบ string: "YYYY-MM-DD HH:mm:ss" เท่านั้น

กฎสำคัญสำหรับการแปลง trans_time (อ้างอิงปีปัจจุบัน พ.ศ. 2569 / ค.ศ. 2026):
1. หากสลิปแสดงปีเป็น พ.ศ. ให้แปลงเป็น ค.ศ. ก่อนเสมอ โดยการลบด้วย 543 (เช่น ปี 2569 ต้องแปลงเป็น 2026, ปี 2568 แปลงเป็น 2025)
2. รูปแบบผลลัพธ์ของเวลาต้องเป็น "YYYY-MM-DD HH:mm:ss" หากในสลิปไม่มีวินาที ให้เติม ":00" ต่อท้ายอัตโนมัติ

ศึกษาและปฏิบัติตามตัวอย่างการแปลงเวลาด้านล่างนี้:
- สลิปแสดง: "6 ก.ค. 69 18:30 น."   -> trans_time: "2026-07-06 18:30:00"
- สลิปแสดง: "06 ก.ค. 2569 18:30"   -> trans_time: "2026-07-06 18:30:00"
- สลิปแสดง: "06 Jul 2026, 18:30:15" -> trans_time: "2026-07-06 18:30:15"
- สลิปแสดง: "2026/07/06 18:30"      -> trans_time: "2026-07-06 18:30:00"
กฎสำคัญ:
1. ถ้าฟิลด์ไหนอ่านไม่ออกหรือไม่มีในภาพ ให้ตอบค่าว่าง ("" สำหรับ string หรือ 0 สำหรับ number)
   ห้ามเดาหรือแต่งข้อมูลขึ้นเองโดยเด็ดขาด
2. amount ต้องเป็น number ไม่ใช่ string`

// slipResponseSchema บังคับ shape ของ JSON ที่ Gemini ต้องตอบกลับมา
// ทำให้ไม่ต้องเขียน parser/regex เดา format เอง
var slipResponseSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"amount":        {Type: genai.TypeNumber},
		"sender_name":   {Type: genai.TypeString},
		"receiver_name": {Type: genai.TypeString},
		"trans_time":    {Type: genai.TypeString},
	},
	Required: []string{"amount", "sender_name", "receiver_name", "trans_time"},
}

// ExtractData ส่งภาพสลิปไปให้ Gemini อ่าน แล้วแปลงผลลัพธ์เป็น domain.GeminiSlipData

func (c *Client) ExtractData(ctx context.Context, imageBytes []byte) (*domain.GeminiSlipData, error) {
	//ถ้าตั้งค่าสูงๆ (เช่น 1.0) AI จะตอบแบบมีความคิดสร้างสรรค์ มีจินตนาการ เดาคำศัพท์ใหม่ๆ
	//ถ้าตั้งค่าต่ำสุดคือ 0.0 AI จะตัดความสร้างสรรค์ออกทั้งหมด แล้วเลือกตอบคำถามตามข้อเท็จจริงที่ตาเห็นจากภาพ 100% (Deterministic) ซึ่งจำเป็นมากๆ สำหรับงานอ่านเลขสลิปธนาคารเพื่อไม่ให้ AI มโนตัวเลขขึ้นมาเอง
	temp := float32(0.0)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   slipResponseSchema,
		Temperature:      &temp,
	}

	contents := []*genai.Content{
		{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				{Text: slipScanPrompt},
				{InlineData: &genai.Blob{MIMEType: "image/jpeg", Data: imageBytes}},
			},
		},
	}

	resp, err := c.genaiClient.Models.GenerateContent(ctx, c.modelName, contents, config)
	if err != nil {
		errStr := err.Error()

		// ตรวจสอบว่า error ที่เกิดขึ้นเป็นปัญหาเกี่ยวกับ Quota หรือไม่
		if strings.Contains(errStr, "RESOURCE_EXHAUSTED") ||
			strings.Contains(errStr, "QUOTA_LIMIT_EXCEEDED") ||
			strings.Contains(errStr, "credits are depleted") ||
			strings.Contains(errStr, "429") {
			return nil, fmt.Errorf("%w: %v", domain.ErrGeminiQuotaExhausted, err)
		}

		// Error อื่นๆ จาก Gemini
		return nil, fmt.Errorf("%w: %v", domain.ErrGeminiUnavailable, err)
	}

	if resp == nil || resp.Text() == "" {
		return nil, domain.ErrGeminiEmptyResponse
	}

	var data domain.GeminiSlipData
	if err := json.Unmarshal([]byte(resp.Text()), &data); err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrSlipParseFailed, err)
	}

	return &data, nil
}
