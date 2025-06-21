package variable

import (
	"context"

	"go.uber.org/zap"
)

func CreateBaseVariables(ctx context.Context, logger *zap.Logger, varRepo IVariableRepo) error {
	modVar := Variable{ID: "back_to_now_agreement", Vars: map[string]any{"page": 1, "signed_at": "2011-02-02T00:00:00Z"}}
	_, err := varRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = varRepo.Create(ctx, &modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
			return err
		}
	}
	modVar = Variable{ID: "back_to_now_contract40000", Vars: map[string]any{"page": 1, "signed_at": "2020-02-02T00:00:00Z", "fz": "fz44", "price_from": 0, "price_to": 40000}}
	_, err = varRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = varRepo.Create(ctx, &modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
			return err
		}
	}
	modVar = Variable{ID: "back_to_now_contract100000", Vars: map[string]any{"page": 1, "signed_at": "2020-02-02T00:00:00Z", "fz": "fz44", "price_from": 40000, "price_to": 100000}}
	_, err = varRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = varRepo.Create(ctx, &modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
			return err
		}
	}
	modVar = Variable{ID: "back_to_now_contract300000", Vars: map[string]any{"page": 1, "signed_at": "2020-02-02T00:00:00Z", "fz": "fz44", "price_from": 100000, "price_to": 300000}}
	_, err = varRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = varRepo.Create(ctx, &modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
			return err
		}
	}
	modVar = Variable{ID: "back_to_now_contract600000", Vars: map[string]any{"page": 1, "signed_at": "2020-02-02T00:00:00Z", "fz": "fz44", "price_from": 300000, "price_to": 600000}}
	_, err = varRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = varRepo.Create(ctx, &modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
			return err
		}
	}
	modVar = Variable{ID: "back_to_now_contract10000000", Vars: map[string]any{"page": 1, "signed_at": "2020-02-02T00:00:00Z", "fz": "fz44", "price_from": 600000, "price_to": 10000000}}
	_, err = varRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = varRepo.Create(ctx, &modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
			return err
		}
	}
	modVar = Variable{ID: "back_to_now_contract999999999999", Vars: map[string]any{"page": 1, "signed_at": "2020-02-02T00:00:00Z", "fz": "fz44", "price_from": 10000000, "price_to": 999999999999}}
	_, err = varRepo.GetByID(ctx, modVar.ID)
	if err != nil {
		err = varRepo.Create(ctx, &modVar)
		if err != nil {
			logger.Error("Ошибка в создании базовы переменных " + err.Error())
			return err
		}
	}
	return nil

}
