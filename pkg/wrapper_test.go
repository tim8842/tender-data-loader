package pkg_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/pkg"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// Mock для интерфейса FuncInWrapp
type mockFunc struct {
	mock.Mock
}

func (m *mockFunc) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	args := m.Called(ctx, logger)
	return args.Get(0), args.Error(1)
}

func TestFuncWrapper_SuccessFirstTry(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	m := new(mockFunc)
	m.On("Process", ctx, logger).Return("success", nil).Once()

	res, err := pkg.FuncWrapper(ctx, logger, 3, 10*time.Millisecond, m)

	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	m.AssertNumberOfCalls(t, "Process", 1)
}

func TestFuncWrapper_RetryThenSuccess(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	m := new(mockFunc)
	// Первые две попытки — ошибка, третья — успех
	m.On("Process", ctx, logger).Return(nil, errors.New("fail 1")).Once()
	m.On("Process", ctx, logger).Return(nil, errors.New("fail 2")).Once()
	m.On("Process", ctx, logger).Return("success after retries", nil).Once()

	res, err := pkg.FuncWrapper(ctx, logger, 3, 10*time.Millisecond, m)

	assert.NoError(t, err)
	assert.Equal(t, "success after retries", res)
	m.AssertNumberOfCalls(t, "Process", 3)
}

func TestFuncWrapper_ExceedMaxRetries(t *testing.T) {
	ctx := context.Background()
	logger := zaptest.NewLogger(t)

	m := new(mockFunc)
	// Все попытки возвращают ошибку
	m.On("Process", ctx, logger).Return(nil, errors.New("fail")).Times(4) // maxRetries=3, значит 4 вызова

	res, err := pkg.FuncWrapper(ctx, logger, 3, 10*time.Millisecond, m)

	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute function after 3 retries")
	m.AssertNumberOfCalls(t, "Process", 4)
}
