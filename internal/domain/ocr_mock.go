package domain

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type OCRGatewayMock struct {
	mock.Mock
}

func (m *OCRGatewayMock) Extract(ctx context.Context, imageByte []byte) (*OCRData, error) {
	args := m.Called(ctx, imageByte)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OCRData), args.Error(1)
}
