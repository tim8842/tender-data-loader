package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/repository"
)

type MockGenericRepository[T repository.BaseModel] struct {
	mock.Mock
}

func (m *MockGenericRepository[T]) GetByID(ctx context.Context, id string) (T, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(T), args.Error(1)
}
func (m *MockGenericRepository[T]) BulkMergeMany(ctx context.Context, docs []T) error {
	args := m.Called(ctx, docs)
	return args.Error(0)
}

func (m *MockGenericRepository[T]) Update(ctx context.Context, id string, data T) error {
	args := m.Called(ctx, id, data)
	return args.Error(0)
}

func (m *MockGenericRepository[T]) Create(ctx context.Context, doc T) error {
	args := m.Called(ctx, doc)
	return args.Error(0)
}
