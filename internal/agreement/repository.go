package agreement

import (
	"context"

	"github.com/tim8842/tender-data-loader/pkg/repository"
)

type AgreementRepo struct {
	*repository.GenericRepository[*Agreement]
}

func (r *AgreementRepo) GetByID(ctx context.Context, id string) (*Agreement, error) {
	return r.GenericRepository.GetByID(ctx, id)
}

func (r *AgreementRepo) BulkMergeMany(ctx context.Context, docs []*Agreement) error {
	return r.BulkCreateOrUpdateMany(ctx, docs)
}

type IAgreementRepo interface {
	GetByID(ctx context.Context, id string) (*Agreement, error)
	BulkMergeMany(ctx context.Context, docs []*Agreement) error
}
