package db

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"go.uber.org/zap"
)

func CreateBase(ctx context.Context, variableRepo *repository.MongoRepository[model.Variable], logger *zap.Logger) error {
	modVar := model.Variable{ID: "back_to_now_agreement", Vars: map[string]any{"page": 50, "signed_at": "2011-02-02T00:00:00Z"}}
	_, err := variableRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = variableRepo.Create(ctx, modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
		}
		return err
	}
	return nil

}
