package domain

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type GeminiSlipRepositoryMock struct {
	mock.Mock
}

func (m *GeminiSlipRepositoryMock) ExtractData(ctx context.Context, imageByte []byte) (*GeminiSlipData, error) {
	args := m.Called(ctx, imageByte)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*GeminiSlipData), args.Error(1)
}
