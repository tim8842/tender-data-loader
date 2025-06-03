package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"
)

type MongoRepository[T any] struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewRepository[T any](collection *mongo.Collection, logger *zap.Logger) *MongoRepository[T] {
	return &MongoRepository[T]{collection: collection, logger: logger}
}

func (r *MongoRepository[T]) Create(ctx context.Context, doc T) error {
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		r.logger.Error("Failed to insert document", zap.Error(err))
	}
	return err
}

func (r *MongoRepository[T]) CreateMany(ctx context.Context, docs []T) error {
	interfaceDocs := make([]interface{}, len(docs))
	for i, d := range docs {
		interfaceDocs[i] = d
	}

	_, err := r.collection.InsertMany(ctx, interfaceDocs)
	if err != nil {
		r.logger.Error("Failed to insert documents", zap.Error(err))
	}
	return err
}

func (r *MongoRepository[T]) GetByID(ctx context.Context, id string) (T, error) {
	var result T
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		r.logger.Error("Failed to get document by ID", zap.String("id", id), zap.Error(err))
	}
	return result, err
}

func (r *MongoRepository[T]) Update(ctx context.Context, id string, update interface{}) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		r.logger.Error("Failed to update document", zap.String("id", id), zap.Error(err))
	}
	return err
}

func (r *MongoRepository[T]) Delete(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to delete document", zap.String("id", id), zap.Error(err))
	}
	return err
}

func (r *MongoRepository[T]) List(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]T, error) {
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
	}
	return results, err
}
