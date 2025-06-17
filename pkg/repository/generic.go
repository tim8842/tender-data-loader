package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type GenericRepository[T BaseModel] struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

type BaseModel interface {
	GetID() any
}

// NewGenericRepository constructs a new GenericRepository[T]
func NewGenericRepository[T BaseModel](coll *mongo.Collection, logger *zap.Logger) *GenericRepository[T] {
	return &GenericRepository[T]{collection: coll, logger: logger}
}

// Create inserts a single document of type T
func (r *GenericRepository[T]) Create(ctx context.Context, doc T) error {
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		r.logger.Error("Failed to insert document", zap.Error(err))
	}
	return err
}

// CreateMany inserts multiple documents of type T
func (r *GenericRepository[T]) CreateMany(ctx context.Context, docs []T) error {
	if len(docs) == 0 {
		return nil
	}
	iface := make([]interface{}, len(docs))
	for i, v := range docs {
		iface[i] = v
	}
	_, err := r.collection.InsertMany(ctx, iface)
	if err != nil {
		r.logger.Error("Failed to insert documents", zap.Error(err))
	}
	return err
}

func (r *GenericRepository[T]) BulkCreateOrUpdateMany(ctx context.Context, docs []T) error {
	if len(docs) == 0 {
		return nil
	}

	var models []mongo.WriteModel
	for _, doc := range docs {
		filter := bson.M{"_id": doc.GetID()}
		replaceModel := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(doc).
			SetUpsert(true)

		models = append(models, replaceModel)
	}
	_, err := r.collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		r.logger.Error("Failed to perform bulk upsert operations", zap.Error(err))
		return err
	}
	return nil
}

// GetByID retrieves a document by its ID and decodes into T
func (r *GenericRepository[T]) GetByID(ctx context.Context, id string) (T, error) {
	var result T
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		r.logger.Error("Failed to get document by ID", zap.String("id", id), zap.Error(err))
		var zero T
		return zero, err
	}
	return result, nil
}

// Update updates a document by its ID
func (r *GenericRepository[T]) Update(ctx context.Context, id string, update T) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		r.logger.Error("Failed to update document", zap.String("id", id), zap.Error(err))
	}
	return err
}

// Delete removes a document by its ID
func (r *GenericRepository[T]) Delete(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to delete document", zap.String("id", id), zap.Error(err))
	}
	return err
}

// List finds documents matching a filter and decodes into []T
func (r *GenericRepository[T]) List(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		r.logger.Error("Failed to list documents", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []T
	for cursor.Next(ctx) {
		var elem T
		if err := cursor.Decode(&elem); err != nil {
			r.logger.Error("Failed to decode document from cursor", zap.Error(err))
			return nil, err
		}
		results = append(results, elem)
	}
	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error after iteration", zap.Error(err))
		return results, err
	}
	return results, nil
}

func (r *GenericRepository[T]) CountDocuments(ctx context.Context, filter interface{}) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *GenericRepository[T]) ReturnCollection() *mongo.Collection {
	return r.collection
}
