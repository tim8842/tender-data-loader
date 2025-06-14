package repository

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/model"
)

type IAgreementRepo interface {
	GetByID(ctx context.Context, id string) (*model.Agreement, error)
	BulkMergeMany(ctx context.Context, docs []*model.Agreement) error
}

type AgreementRepo struct {
	*GenericRepository[*model.Agreement]
}

func (r *AgreementRepo) GetByID(ctx context.Context, id string) (*model.Agreement, error) {
	return r.GenericRepository.GetByID(ctx, id)
}

func (r *AgreementRepo) BulkMergeMany(ctx context.Context, docs []*model.Agreement) error {
	return r.BulkCreateOrUpdateMany(ctx, docs)
}
