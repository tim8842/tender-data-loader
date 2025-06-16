package mongo

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/mocks"
	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"go.uber.org/zap"
)

func TestCreateBase(t *testing.T) {

	type testCase struct {
		name        string
		expectedErr bool
	}
	tests := []testCase{
		{
			name:        "Bad create back_to_now_agreement",
			expectedErr: true,
		},
	}
	ctx := context.Background()
	logger := zap.NewNop()
	mockVarR := new(mocks.MockGenericRepository[*model.Variable])
	modVar := model.Variable{ID: "back_to_now_agreement", Vars: map[string]any{"page": 50, "signed_at": "2011-02-02T00:00:00Z"}}
	reps := &repository.Repositories{VarRepo: mockVarR}
	for _, tt := range tests {
		mockVarR.On("GetByID", mock.Anything, "back_to_now_agreement").Return(&model.Variable{}, errors.New("")).Once()
		mockVarR.On("Create", mock.Anything, &modVar).Return(errors.New("")).Once()
		res := CreateBase(ctx, logger, reps)
		if tt.expectedErr {
			assert.Error(t, res)
		} else {
			assert.NoError(t, res)
		}
	}
}
