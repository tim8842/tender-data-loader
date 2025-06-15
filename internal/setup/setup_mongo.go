package setup

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"github.com/tim8842/tender-data-loader/pkg/db"
)

type MongoConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
}

func SetupMongo(ctx context.Context, logger *zap.Logger, cfg *MongoConfig) (*mongo.Client, *mongo.Database, error) {

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s/?authSource=admin", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	client, err := db.ConnectMongo(ctx, logger, uri)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка подключения к MongoDB: %w", err)
	}

	db := client.Database(cfg.DBName)
	return client, db, nil
}
