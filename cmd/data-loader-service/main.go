package main

import (
	"context"
	"log"
	"time"

	"github.com/tim8842/tender-data-loader/internal/config"
	loggerPackage "github.com/tim8842/tender-data-loader/internal/logger"
	"github.com/tim8842/tender-data-loader/internal/model"
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
	logger, _, err := loggerPackage.InitLogger("./logs", 100, 7, 30, true)
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
		&setup.MongoConfig{
			User: cfg.MongoUser, Password: cfg.MongoPassword,
			Host: cfg.MongoHost, Port: cfg.MongoPort,
			DBName: cfg.MongoDB,
		},
	)
	if err != nil {
		logger.Fatal("Ошибка подключеник к монго", zap.Error(err))
	}
	defer client.Disconnect(mainCtx)
	genAgreeRepo := repository.NewGenericRepository[*model.Agreement](dbConn.Collection("agreements"), logger)
	agreementRepo := &repository.AgreementRepo{GenericRepository: genAgreeRepo}
	genVarRepo := repository.NewGenericRepository[*model.Variable](dbConn.Collection("variables"), logger)
	variableRepo := &repository.VariableRepo{GenericRepository: genVarRepo}
	genCustomerRepo := repository.NewGenericRepository[*model.Customer](dbConn.Collection("customers"), logger)
	customerRepo := &repository.CustomerRepo{GenericRepository: genCustomerRepo}
	repositories := &repository.Repositories{AgreementRepo: agreementRepo, CustomerRepo: customerRepo, VarRepo: variableRepo}
	dbp.CreateBase(ctxTimeout, logger, repositories)
	go tasks.StartTasks(mainCtx, logger, cfg, repositories)
	app := setup.SetupFiberApp(repositories, logger)

	logger.Info("Сервер запущен на :8080")
	if err := app.Listen(":8080"); err != nil {
		logger.Fatal("Ошибка при запуске сервера", zap.Error(err))
	}
}
