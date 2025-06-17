package pkg_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/pkg"
	"go.uber.org/zap"
)

// MockTaskHandler — мок для TaskHandler
type MockTaskHandler struct {
	mock.Mock
}

func (m *MockTaskHandler) Process(ctx context.Context, logger *zap.Logger) error {
	args := m.Called(ctx, logger)
	return args.Error(0)
}

// newTestLogger возвращает zap.Logger, который пишет в тест
func newTestLogger(t *testing.T) *zap.Logger {
	cfg := zap.NewDevelopmentConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg.OutputPaths = []string{"stdout"}
	cfg.EncoderConfig.TimeKey = "" // Упростим логи

	logger, err := cfg.Build()
	if err != nil {
		t.Fatalf("failed to build logger: %v", err)
	}
	return logger
}

func TestTaskRunner(t *testing.T) {
	logger := newTestLogger(t)

	tests := []struct {
		name          string
		setupMocks    func() map[string]*MockTaskHandler
		taskNames     []string
		expectedCalls map[string]int
		expectedErrs  map[string]error
	}{
		{
			name: "single_successful_task",
			setupMocks: func() map[string]*MockTaskHandler {
				m := &MockTaskHandler{}
				m.On("Process", mock.Anything, mock.Anything).Return(nil).Once()
				return map[string]*MockTaskHandler{"task1": m}
			},
			taskNames:     []string{"task1"},
			expectedCalls: map[string]int{"task1": 1},
		},
		{
			name: "task_with_error",
			setupMocks: func() map[string]*MockTaskHandler {
				m := &MockTaskHandler{}
				m.On("Process", mock.Anything, mock.Anything).Return(errors.New("failed")).Once()
				return map[string]*MockTaskHandler{"task2": m}
			},
			taskNames:     []string{"task2"},
			expectedCalls: map[string]int{"task2": 1},
		},
		{
			name: "unknown_task_name",
			setupMocks: func() map[string]*MockTaskHandler {
				return map[string]*MockTaskHandler{}
			},
			taskNames:     []string{"unknown"},
			expectedCalls: map[string]int{},
		},
		{
			name: "multiple_tasks",
			setupMocks: func() map[string]*MockTaskHandler {
				mA := &MockTaskHandler{}
				mA.On("Process", mock.Anything, mock.Anything).Return(nil).Times(2)

				mB := &MockTaskHandler{}
				mB.On("Process", mock.Anything, mock.Anything).Return(nil).Once()

				return map[string]*MockTaskHandler{
					"taskA": mA,
					"taskB": mB,
				}
			},
			taskNames:     []string{"taskA", "taskB", "taskA"},
			expectedCalls: map[string]int{"taskA": 2, "taskB": 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			runner := pkg.NewTaskRunner(ctx, logger, 1)
			mocks := tt.setupMocks()

			for name, mockHandler := range mocks {
				runner.RegisterTask(name, mockHandler)
			}

			runner.Start()

			// Подождать, чтобы воркер успел запуститься
			time.Sleep(50 * time.Millisecond)

			for _, taskName := range tt.taskNames {
				runner.Enqueue(taskName)
			}

			// Подождать, чтобы таски были обработаны
			time.Sleep(100 * time.Millisecond)

			runner.Stop()

			for name, mockHandler := range mocks {
				mockHandler.AssertNumberOfCalls(t, "Process", tt.expectedCalls[name])
			}
		})
	}
}
