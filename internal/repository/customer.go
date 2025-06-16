package repository

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/model"
)

type CustomerRepo struct {
	*GenericRepository[*model.Customer]
}

func (r *CustomerRepo) BulkMergeMany(ctx context.Context, docs []*model.Customer) error {
	return r.BulkCreateOrUpdateMany(ctx, docs)
}
