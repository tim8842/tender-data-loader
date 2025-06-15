package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/logger"
	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"github.com/tim8842/tender-data-loader/internal/setup"
	"github.com/tim8842/tender-data-loader/internal/tasks"
	"github.com/tim8842/tender-data-loader/pkg/db/mongo"
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
	lgr, _, err := logger.InitLogger("./logs", 100, 7, 30, true)
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer lgr.Sync()
	lgr.Info("Логгер инициализирован")
	// Загрузка конфига
	envPath, err := filepath.Abs(".env")
	if err != nil {
		panic(err)
	}
	cfg, err := config.LoadConfig(envPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	client, dbConn, err := setup.SetupMongo(
		ctxTimeout,
		lgr,
		&setup.MongoConfig{
			User: cfg.MongoUser, Password: cfg.MongoPassword,
			Host: cfg.MongoHost, Port: cfg.MongoPort,
			DBName: cfg.MongoDB,
		},
	)
	if err != nil {
		log.Fatal("Ошибка подключеник к монго", zap.Error(err))
	}
	defer client.Disconnect(mainCtx)
	genAgreeRepo := repository.NewGenericRepository[*model.Agreement](dbConn.Collection("agreements"), lgr)
	agreementRepo := &repository.AgreementRepo{GenericRepository: genAgreeRepo}
	genVarRepo := repository.NewGenericRepository[*model.Variable](dbConn.Collection("variables"), lgr)
	variableRepo := &repository.VariableRepo{GenericRepository: genVarRepo}
	genCustomerRepo := repository.NewGenericRepository[*model.Customer](dbConn.Collection("customers"), lgr)
	customerRepo := &repository.CustomerRepo{GenericRepository: genCustomerRepo}
	repositories := &repository.Repositories{AgreementRepo: agreementRepo, CustomerRepo: customerRepo, VarRepo: variableRepo}
	mongo.CreateBase(ctxTimeout, lgr, repositories)
	go tasks.StartTasks(mainCtx, lgr, cfg, repositories)
	app := setup.SetupFiberApp(repositories, lgr)

	lgr.Info("Сервер запущен на :" + cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		lgr.Fatal("Ошибка при запуске сервера", zap.Error(err))
	}
}
