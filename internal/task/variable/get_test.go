package variable_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	inmock "github.com/tim8842/tender-data-loader/internal/mock"
	task "github.com/tim8842/tender-data-loader/internal/task/variable"
	"github.com/tim8842/tender-data-loader/internal/variable"

	"github.com/tim8842/tender-data-loader/pkg/parser"
	"go.uber.org/zap/zaptest"
)

func TestGetVariableBackToNowById_Process(t *testing.T) {
	date, _ := parser.ParseFromDateToTime("12.12.2024")
	tests := []struct {
		name      string
		mockData  *variable.Variable
		mockError error
		expected  *variable.VariableBackToNow
		expectErr bool
	}{
		{
			name: "success",
			mockData: &variable.Variable{
				ID:   "back_to",
				Vars: map[string]any{"page": 1, "signed_at": date},
			},
			mockError: nil,
			expected: &variable.VariableBackToNow{
				ID: "back_to",
				Vars: variable.VarsBackToNow{
					Page:     1,
					SignedAt: date,
				},
			},
			expectErr: false,
		},
		{
			name:      "repo returns error",
			mockData:  nil,
			mockError: errors.New("db error"),
			expected:  nil,
			expectErr: true,
		},
		{
			name: "unmarshal error",
			mockData: &variable.Variable{
				ID:   "back_to",
				Vars: map[string]any{"page": 1, "signed_yat": date, "enbat": "dadas"},
			},
			mockError: errors.New("marshal error"),
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(inmock.MockGenericRepository[*variable.Variable])
			ctx := context.Background()
			logger := zaptest.NewLogger(t)

			mockRepo.On("GetByID", mock.Anything, "id123").Return(tt.mockData, tt.mockError)

			task := task.NewGetVariableBackToNowById(mockRepo, "id123")
			result, err := task.Process(ctx, logger)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
