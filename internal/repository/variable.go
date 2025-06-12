package repository

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/model"
)

type IVariableRepo interface {
	GetByID(ctx context.Context, id string) (*model.Variable, error)
	Update(ctx context.Context, id string, data *model.Variable) error
	Create(ctx context.Context, doc *model.Variable) error
}

type VariableRepo struct {
	*GenericRepository[*model.Variable]
}

func (r *VariableRepo) GetByID(ctx context.Context, id string) (*model.Variable, error) {
	return r.GenericRepository.GetByID(ctx, id)

}

func (r *VariableRepo) Update(ctx context.Context, id string, data *model.Variable) error {
	return r.GenericRepository.Update(ctx, id, data)

}

func (r *VariableRepo) Create(ctx context.Context, doc *model.Variable) error {
	return r.GenericRepository.Create(ctx, doc)

}
