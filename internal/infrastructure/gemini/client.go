package gemini

import (
	"context"
	"fmt"

	"github.com/Jaruvat303/cashlog/cmd/config"
	"google.golang.org/genai"
)

// Client ครอบ genai.Client ของ Gemini SDK ไว้อีกชั้น
// พร้อมเก็บชื่อโมเดลที่จะใช้เรียก

type Client struct {
	genaiClient *genai.Client
	modelName   string
}

// NewClient เชื่อมต่อกับ Gemini API และคืนค่า Client ที่พร้อมใช้งาน
//
// เรียกครั้งเดียวตอน bootstrap แอป (เช่นใน main.go) แล้วส่ง instance
// เดียวกันไปใช้ตลอดอายุของโปรแกรม ไม่ต้องสร้างใหม่ทุก request
func NewClient(ctx context.Context, cfg config.Config) (*Client, error) {
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("gemini: api key is required")
	}

	modelName := cfg.ModelName
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GeminiAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}

	return &Client{
		genaiClient: c,
		modelName:   modelName,
	}, nil
}
