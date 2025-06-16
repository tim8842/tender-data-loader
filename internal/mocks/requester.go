package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/util/requests"
	"go.uber.org/zap"
)

type MockRequester struct {
	mock.Mock
}

func (m *MockRequester) Get(ctx context.Context, logger *zap.Logger, url string, timeout time.Duration, opts ...*requests.RequestOptions) ([]byte, error) {
	args := m.Called(ctx, logger, url, timeout, opts)
	// Приведение к []byte обязательно
	return args.Get(0).([]byte), args.Error(1)
}
