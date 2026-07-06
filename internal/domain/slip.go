package domain

import (
	"context"
)

type GeminiSlipData struct {
	Amount       float64 `json:"amount"`
	SenderName   string  `json:"sender_name"`
	ReceiverName string  `json:"receiver_name"`
	TransTime    string  `json:"trans_time"`
}

type GeminiSlipRepository interface {
	ExtractData(ctx context.Context, imageData []byte) (*GeminiSlipData, error)
}
