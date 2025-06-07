package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/tim8842/tender-data-loader/internal/config"
	loggerPackage "github.com/tim8842/tender-data-loader/internal/logger"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"github.com/tim8842/tender-data-loader/internal/setup"
	"github.com/tim8842/tender-data-loader/internal/tasks"
	dbp "github.com/tim8842/tender-data-loader/pkg/db"
	"go.uber.org/zap"
)

// @title Tender API
// @version 1.0
// @description API для работы с договорами
// @host localhost:8080
// @BasePath /

func main() {
	mainCtx, cancel := context.WithCancel(context.Background())
	//закрываю родителя, должны закрыться и дети
	ctxTimeout, cancelTimeout := context.WithTimeout(mainCtx, 10*time.Second)
	defer cancel()
	defer cancelTimeout()
	// Загрузка логера
	_ = godotenv.Load()
	logger, err := loggerPackage.InitLogger("./logs")
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer logger.Sync()
	logger.Info("Логгер инициализирован")
	// Загрузка конфига
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	client, dbConn, err := setup.SetupMongo(
		ctxTimeout,
		logger,
		setup.MongoConfig{
			User: cfg.MongoUser, Password: cfg.MongoPassword,
			Host: cfg.MongoHost, Port: cfg.MongoPort,
			DBName: cfg.MongoDB,
		},
	)
	if err != nil {
		logger.Fatal("Ошибка подключеник к монго", zap.Error(err))
	}
	defer client.Disconnect(mainCtx)

	agreementRepo := repository.NewRepository(dbConn.Collection("agreements"), logger)
	variableRepo := repository.NewRepository(dbConn.Collection("variables"), logger)
	dbp.CreateBase(ctxTimeout, variableRepo, logger)
	go tasks.StartTasks(mainCtx, logger, variableRepo)
	app := setup.SetupFiberApp(&repository.Repositories{AgreementRepo: agreementRepo}, logger)

	logger.Info("Сервер запущен на :8080")
	if err := app.Listen(":8080"); err != nil {
		logger.Fatal("Ошибка при запуске сервера", zap.Error(err))
	}
}
