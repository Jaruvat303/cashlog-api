package gemini

import (
	"context"
	"encoding/json"
	"fmt"

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
- trans_time: วันและเวลาที่ทำรายการ แปลงเป็นรูปแบบ string: "2006-01-02 15:04:05"
  ถ้าสลิปแสดงเป็น พ.ศ. ให้แปลงเป็น ค.ศ. ก่อน (พ.ศ. - 543)
 
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

// ScanSlip ส่งภาพสลิปไปให้ Gemini อ่าน แล้วแปลงผลลัพธ์เป็น domain.GeminiSlipData
//
// method นี้คือ implementation ของ domain.SlipScanner interface
// นั่นคือเหตุผลที่ signature ต้องตรงกับ interface เป๊ะๆ
func (c *Client) ExtractData(ctx context.Context, imageBytes []byte) (*domain.GeminiSlipData, error) {
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   slipResponseSchema,
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
		return nil, fmt.Errorf("gemini: generate content failed: %w", err)
	}

	if resp == nil || resp.Text() == "" {
		return nil, fmt.Errorf("gemini: empty response from model")
	}

	var data domain.GeminiSlipData
	if err := json.Unmarshal([]byte(resp.Text()), &data); err != nil {
		return nil, fmt.Errorf("gemini: failed to parse response json: %w", err)
	}

	// validate เบื้องต้นตรงนี้ก่อนส่งกลับไปให้ usecase
	// เพื่อกัน case ที่ Gemini ตอบ JSON ถูก schema แต่ค่าไม่สมเหตุสมผล
	if data.Amount <= 0 {
		return nil, fmt.Errorf("gemini: invalid amount extracted: %v", data.Amount)
	}

	return &data, nil
}
