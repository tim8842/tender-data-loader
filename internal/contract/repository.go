package contract

import (
	"context"

	"github.com/tim8842/tender-data-loader/pkg/repository"
)

type ContractRepo struct {
	*repository.GenericRepository[*Contract]
}

func (r *ContractRepo) GetByID(ctx context.Context, id string) (*Contract, error) {
	return r.GenericRepository.GetByID(ctx, id)
}

func (r *ContractRepo) BulkMergeMany(ctx context.Context, docs []*Contract) error {
	return r.BulkCreateOrUpdateMany(ctx, docs)
}

type IContractRepo interface {
	GetByID(ctx context.Context, id string) (*Contract, error)
	BulkMergeMany(ctx context.Context, docs []*Contract) error
}
