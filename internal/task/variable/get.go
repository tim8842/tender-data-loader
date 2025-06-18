package variable

import (
	"context"
	"encoding/json"

	"github.com/tim8842/tender-data-loader/internal/variable"
	"go.uber.org/zap"
)

type GetVariableBackToNowById struct {
	variableRepo variable.IVariableRepo
	id           string
}

func NewGetVariableBackToNowById(variableRepo variable.IVariableRepo, id string) *GetVariableBackToNowById {
	return &GetVariableBackToNowById{variableRepo: variableRepo, id: id}
}

func (t GetVariableBackToNowById) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := t.variableRepo.GetByID(ctx, t.id)
	if ok != nil {
		return nil, ok
	}
	var model variable.VariableBackToNow
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
