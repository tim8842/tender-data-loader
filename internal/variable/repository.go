package variable

import (
	"context"

	"github.com/tim8842/tender-data-loader/pkg/repository"
)

type VariableRepo struct {
	*repository.GenericRepository[*Variable]
}

func (r *VariableRepo) GetByID(ctx context.Context, id string) (*Variable, error) {
	return r.GenericRepository.GetByID(ctx, id)

}

func (r *VariableRepo) Update(ctx context.Context, id string, data *Variable) error {
	return r.GenericRepository.Update(ctx, id, data)

}

func (r *VariableRepo) Create(ctx context.Context, doc *Variable) error {
	return r.GenericRepository.Create(ctx, doc)

}

type IVariableRepo interface {
	GetByID(ctx context.Context, id string) (*Variable, error)
	Update(ctx context.Context, id string, data *Variable) error
	Create(ctx context.Context, doc *Variable) error
}
