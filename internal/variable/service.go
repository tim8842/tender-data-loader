package variable

import (
	"context"

	"go.uber.org/zap"
)

func CreateBaseVariables(ctx context.Context, logger *zap.Logger, varRepo IVariableRepo) error {
	modVar := Variable{ID: "back_to_now_agreement", Vars: map[string]any{"page": 50, "signed_at": "2011-02-02T00:00:00Z"}}
	_, err := varRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = varRepo.Create(ctx, &modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
		}
		return err
	}
	return nil

}
