package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/fiber"
	"github.com/tim8842/tender-data-loader/internal/task"
	"github.com/tim8842/tender-data-loader/internal/variable"
	"github.com/tim8842/tender-data-loader/pkg/db/mongo"
	"github.com/tim8842/tender-data-loader/pkg/logger"
	"github.com/tim8842/tender-data-loader/pkg/repository"
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
	envPath, err := filepath.Abs("configs/.env")
	if err != nil {
		panic(err)
	}
	cfg, err := config.LoadConfig(envPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	client, dbConn, err := mongo.SetupMongo(
		ctxTimeout,
		lgr,
		&mongo.MongoConfig{
			User: cfg.MongoUser, Password: cfg.MongoPassword,
			Host: cfg.MongoHost, Port: cfg.MongoPort,
			DBName: cfg.MongoDB,
		},
	)
	if err != nil {
		log.Fatal("Ошибка подключеник к монго", zap.Error(err))
	}
	defer client.Disconnect(mainCtx)
	genAgreeRepo := repository.NewGenericRepository[*agreement.Agreement](dbConn.Collection("agreements"), lgr)
	agreementRepo := &agreement.AgreementRepo{GenericRepository: genAgreeRepo}
	genVarRepo := repository.NewGenericRepository[*variable.Variable](dbConn.Collection("variables"), lgr)
	variableRepo := &variable.VariableRepo{GenericRepository: genVarRepo}
	genCustomerRepo := repository.NewGenericRepository[*customer.Customer](dbConn.Collection("customers"), lgr)
	customerRepo := &customer.CustomerRepo{GenericRepository: genCustomerRepo}
	variable.CreateBaseVariables(ctxTimeout, lgr, variableRepo)
	go task.StartTasks(mainCtx, lgr, cfg, agreementRepo, variableRepo, customerRepo)
	app := fiber.SetupFiberApp(lgr, agreementRepo)

	lgr.Info("Сервер запущен на :" + cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		lgr.Fatal("Ошибка при запуске сервера", zap.Error(err))
	}
}
