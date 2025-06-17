package variable_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	inmock "github.com/tim8842/tender-data-loader/internal/mock"
	"github.com/tim8842/tender-data-loader/internal/variable"
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
	mockVarR := new(inmock.MockGenericRepository[*variable.Variable])
	modVar := variable.Variable{ID: "back_to_now_agreement", Vars: map[string]any{"page": 50, "signed_at": "2011-02-02T00:00:00Z"}}
	for _, tt := range tests {
		mockVarR.On("GetByID", mock.Anything, "back_to_now_agreement").Return(&variable.Variable{}, errors.New("")).Once()
		mockVarR.On("Create", mock.Anything, &modVar).Return(errors.New("")).Once()
		res := variable.CreateBaseVariables(ctx, logger, mockVarR)
		if tt.expectedErr {
			assert.Error(t, res)
		} else {
			assert.NoError(t, res)
		}
	}
}
