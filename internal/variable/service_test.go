package variable_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	inmock "github.com/tim8842/tender-data-loader/internal/mock"
	"github.com/tim8842/tender-data-loader/internal/variable"
	"go.uber.org/zap"
)

func TestCreateBaseVariables(t *testing.T) {
	tests := []struct {
		name           string
		missingVarIDs  map[string]bool // какие ID должны отсутствовать (GetByID -> err)
		createErrorFor string          // ID, при котором Create вернёт ошибку (если задан)
		wantErr        bool
	}{
		{
			name:          "Все переменные уже есть",
			missingVarIDs: map[string]bool{},
			wantErr:       false,
		},
		{
			name:          "Одна переменная отсутствует, Create успешно",
			missingVarIDs: map[string]bool{"back_to_now_contract100000": true},
			wantErr:       false,
		},
		{
			name:           "Одна переменная отсутствует, Create возвращает ошибку",
			missingVarIDs:  map[string]bool{"back_to_now_contract999999999999": true},
			createErrorFor: "back_to_now_contract999999999999",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(inmock.MockGenericRepository[*variable.Variable])
			logger := zap.NewNop()
			ctx := context.TODO()

			// Массив всех переменных, которые будут создаваться
			baseVars := []variable.Variable{
				{ID: "back_to_now_agreement"},
				{ID: "back_to_now_contract40000"},
				{ID: "back_to_now_contract100000"},
				{ID: "back_to_now_contract300000"},
				{ID: "back_to_now_contract600000"},
				{ID: "back_to_now_contract10000000"},
				{ID: "back_to_now_contract999999999999"},
			}

			for _, v := range baseVars {
				if tt.missingVarIDs[v.ID] {
					mockRepo.On("GetByID", mock.Anything, v.ID).Return((*variable.Variable)(nil), errors.New("not found"))

					// Если Create должен вернуть ошибку — проверяем
					if tt.createErrorFor == v.ID {
						mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(actual *variable.Variable) bool {
							return actual.ID == v.ID
						})).Return(errors.New("create failed")).Once()
					} else {
						mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(actual *variable.Variable) bool {
							return actual.ID == v.ID
						})).Return(nil).Once()
					}
				} else {
					mockRepo.On("GetByID", mock.Anything, v.ID).Return(&v, nil)
				}
			}

			err := variable.CreateBaseVariables(ctx, logger, mockRepo)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBaseVariables() error = %v, wantErr %v", err, tt.wantErr)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
