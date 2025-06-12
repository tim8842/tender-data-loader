package subtasks

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/mocks"
	"github.com/tim8842/tender-data-loader/internal/model"
	baseutils "github.com/tim8842/tender-data-loader/internal/util/base_utils"
	"go.uber.org/zap/zaptest"
)

func TestGetVariableBackToNowAgreementById_Process(t *testing.T) {
	date, _ := baseutils.ParseDate("12.12.2024")
	tests := []struct {
		name      string
		mockData  *model.Variable
		mockError error
		expected  *model.VariableBackToNowAgreement
		expectErr bool
	}{
		{
			name: "success",
			mockData: &model.Variable{
				ID:   "back_to",
				Vars: map[string]any{"page": 1, "signed_at": date},
			},
			mockError: nil,
			expected: &model.VariableBackToNowAgreement{
				ID: "back_to",
				Vars: model.VarsBackToNowAgreement{
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
			mockData: &model.Variable{
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
			mockRepo := new(mocks.MockGenericRepository[*model.Variable])
			ctx := context.Background()
			logger := zaptest.NewLogger(t)

			mockRepo.On("GetByID", mock.Anything, "id123").Return(tt.mockData, tt.mockError)

			task := NewGetVariableBackToNowAgreementById(mockRepo, "id123")
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
