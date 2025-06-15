package db

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/util/wrappers"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.uber.org/zap"
)

type PingFuncInWrapp struct {
	client *mongo.Client
}

func NewPingFuncInWrapp(client *mongo.Client) *PingFuncInWrapp {
	return &PingFuncInWrapp{client: client}
}

func (t PingFuncInWrapp) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	err := t.client.Ping(ctx, nil)
	return nil, err
}

// ConnectMongo принимает uri и logger
func ConnectMongo(ctx context.Context, logger *zap.Logger, uri string) (*mongo.Client, error) {
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	// defer client.Disconnect(ctx)
	if err != nil {
		logger.Error("Failed to connect to MongoDB ", zap.Error(err))
		return nil, err
	}

	_, err = wrappers.FuncWrapper(ctx, logger, 3, 5, NewPingFuncInWrapp(client))
	if err != nil {
		logger.Error("Failed to ping MongoDB ", zap.Error(err))
		return nil, err
	}

	logger.Info("Connected to MongoDB successfully")
	return client, nil
}
