package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// IMongoRepository Interface for MongoDB operations
type IMongoRepository interface {
	Create(ctx context.Context, doc interface{}) error
	CreateMany(ctx context.Context, docs []interface{}) error
	GetByID(ctx context.Context, id string) (interface{}, error)
	Update(ctx context.Context, id string, update interface{}) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]map[string]interface{}, error)
}

// MongoRepository Implementation of IMongoRepository using interface{}
type MongoRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewRepository Constructor for MongoRepository
func NewRepository(collection *mongo.Collection, logger *zap.Logger) IMongoRepository {
	return &MongoRepository{collection: collection, logger: logger}
}

// Create Implements IMongoRepository.Create
func (r *MongoRepository) Create(ctx context.Context, doc interface{}) error {
	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		r.logger.Error("Failed to insert document", zap.Error(err))
	}
	return err
}

// CreateMany Implements IMongoRepository.CreateMany
func (r *MongoRepository) CreateMany(ctx context.Context, docs []interface{}) error {
	_, err := r.collection.InsertMany(ctx, convertToInterfaceSlice(docs))
	if err != nil {
		r.logger.Error("Failed to insert documents", zap.Error(err))
	}
	return err
}

// GetByID Implements IMongoRepository.GetByID
func (r *MongoRepository) GetByID(ctx context.Context, id string) (interface{}, error) {
	var result map[string]interface{}

	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		r.logger.Error("Failed to get document by ID", zap.String("id", id), zap.Error(err))
	}
	return result, err
}

// Update Implements IMongoRepository.Update
func (r *MongoRepository) Update(ctx context.Context, id string, update interface{}) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{"$set": update})
	if err != nil {
		r.logger.Error("Failed to update document", zap.String("id", id), zap.Error(err))
	}
	return err
}

// Delete Implements IMongoRepository.Delete
func (r *MongoRepository) Delete(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to delete document", zap.String("id", id), zap.Error(err))
	}
	return err
}

// List Implements IMongoRepository.List
func (r *MongoRepository) List(ctx context.Context, filter interface{}, opts ...*options.FindOptions) ([]map[string]interface{}, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		r.logger.Error("Failed to list documents", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}

	for cursor.Next(ctx) {
		var elem map[string]interface{}

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

// Helper function to convert []T to []interface{}
func convertToInterfaceSlice[T any](docs []T) []interface{} {
	interfaceDocs := make([]interface{}, len(docs))
	for i, d := range docs {
		interfaceDocs[i] = d
	}
	return interfaceDocs
}
