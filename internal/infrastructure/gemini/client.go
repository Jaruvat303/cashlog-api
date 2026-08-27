package gemini

import (
	"context"
	"fmt"

	"github.com/Jaruvat303/cashlog/pkg/logger"
	"go.uber.org/zap"
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
func NewClient(ctx context.Context, geminiAPIKey string, geminiModelName string, appLogger logger.Logger) (*Client, error) {
	if geminiAPIKey == "" {
		appLogger.Fatal("Configuration 'GEMINI_API_KEY' is required but found empty")
		return nil, fmt.Errorf("gemini: api key is required")
	}

	modelName := geminiModelName
	if geminiModelName == "" {
		appLogger.Warn("Configuration 'GEMINI_MODEL_NAME' is empty, using default model 'gemini-2.5-flash'")
		modelName = "gemini-2.5-flash"
	}

	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: geminiAPIKey,
	})
	if err != nil {
		appLogger.Fatal("Failed to create Gemini client: %v", zap.Error(err))
		return nil, fmt.Errorf("gemini: failed to create client: %w", err)
	}

	return &Client{
		genaiClient: c,
		modelName:   modelName,
	}, nil
}
