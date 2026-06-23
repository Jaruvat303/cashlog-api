package domain

import (
	"context"
	"time"
)

type OCRData struct {
	TransactionID   string
	Amount          float64
	ReceiverName    string
	TransactionDate time.Time
}

type OCRGateway interface {
	Extract(ctx context.Context, imageBytes []byte) (*OCRData, error)
}
