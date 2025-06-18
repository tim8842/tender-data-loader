package supplier

import (
	"context"

	"github.com/tim8842/tender-data-loader/pkg/repository"
)

type SupplierRepo struct {
	*repository.GenericRepository[*Supplier]
}

func (r *SupplierRepo) GetByID(ctx context.Context, id string) (*Supplier, error) {
	return r.GenericRepository.GetByID(ctx, id)
}

func (r *SupplierRepo) BulkMergeMany(ctx context.Context, docs []*Supplier) error {
	return r.BulkCreateOrUpdateMany(ctx, docs)
}

type ISupplierRepo interface {
	GetByID(ctx context.Context, id string) (*Supplier, error)
	BulkMergeMany(ctx context.Context, docs []*Supplier) error
}
