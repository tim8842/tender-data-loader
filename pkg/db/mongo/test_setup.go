package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

func SetupTestMongo(ctx context.Context, logger *zap.Logger, cfg *MongoConfig) (*mongo.Client, *mongo.Database, error) {

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=admin", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	client, err := ConnectMongo(ctx, logger, uri)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка подключения к MongoDB: %w", err)
	}

	db := client.Database("dbtest")
	return client, db, nil
}
