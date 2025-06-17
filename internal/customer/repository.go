package customer

import (
	"context"

	"github.com/tim8842/tender-data-loader/pkg/repository"
)

type CustomerRepo struct {
	*repository.GenericRepository[*Customer]
}

func (r *CustomerRepo) BulkMergeMany(ctx context.Context, docs []*Customer) error {
	return r.BulkCreateOrUpdateMany(ctx, docs)
}

type ICustomerRepo interface {
	BulkMergeMany(ctx context.Context, docs []*Customer) error
}
