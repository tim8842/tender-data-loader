package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"
)

// ConnectMongo принимает uri и logger
func ConnectMongo(uri string, logger *zap.Logger) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Error("Failed to ping MongoDB", zap.Error(err))
		return nil, err
	}

	logger.Info("Connected to MongoDB successfully")
	return client, nil
}
