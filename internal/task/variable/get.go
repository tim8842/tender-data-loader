package variable

import (
	"context"
	"encoding/json"

	"github.com/tim8842/tender-data-loader/internal/variable"
	"go.uber.org/zap"
)

type GetVariableBackToNowAgreementById struct {
	variableRepo variable.IVariableRepo
	id           string
}

func NewGetVariableBackToNowAgreementById(variableRepo variable.IVariableRepo, id string) *GetVariableBackToNowAgreementById {
	return &GetVariableBackToNowAgreementById{variableRepo: variableRepo, id: id}
}

func (t GetVariableBackToNowAgreementById) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := t.variableRepo.GetByID(ctx, t.id)
	if ok != nil {
		return nil, ok
	}
	var model variable.VariableBackToNowAgreement
	b, ok := json.Marshal(data)
	if ok != nil {
		return nil, ok
	}
	ok = json.Unmarshal(b, &model)
	if ok == nil {
		return &model, ok
	} else {
		return nil, ok
	}
}
