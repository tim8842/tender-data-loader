package main

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/tim8842/tender-data-loader/internal/handler"
	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	t "github.com/tim8842/tender-data-loader/internal/tasks"
	"github.com/tim8842/tender-data-loader/internal/util"
	dbp "github.com/tim8842/tender-data-loader/pkg/db"
)

// @title Tender API
// @version 1.0
// @description API для работы с договорами
// @host localhost:8080
// @BasePath /

func startTasks(logger *zap.Logger) {
	WorkerCount := 1
	tasks := make(chan string, 1)
	var wg sync.WaitGroup
	wg.Add(WorkerCount)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handlers := map[string]t.TaskHandler{
		"back_to_now_agreement": t.TaskFunc(t.BackToNowAgreementTask),
	}
	for i := 0; i < WorkerCount; i++ {
		go t.Worker(ctx, i+1, tasks, handlers, &wg, logger)
	}
	tasks <- "back_to_now_agreement"
	logger.Info("Send back_to_now_agreement in logger")
	close(tasks)
	wg.Wait()
	logger.Info("Main: all workers finished")
}

func main() {
	_ = godotenv.Load()
	logger, err := util.InitLogger("./logs")
	if err != nil {
		log.Fatalf("Ошибка инициализации логгера: %v", err)
	}
	defer logger.Sync()
	logger.Info("Логгер инициализирован")

	// Подключение к MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// client, err := mongo.Connect(ctx, options.Client().ApplyURI(
	// 	"mongodb://"+os.Getenv("MONGO_USER")+
	// 		":"+os.Getenv("MONGO_PASSWORD")+"@"+
	// 		os.Getenv("MONGO_HOST")+
	// 		":"+os.Getenv("MONGO_PORT")+"/?authSource=admin",
	// ))
	// if err != nil {
	// 	logger.Fatal("Не удалось создать Mongo client", zap.Error(err))
	// }
	defer cancel()
	// defer client.Disconnect(ctx)
	client, _ := dbp.ConnectMongo(ctx, "mongodb://"+os.Getenv("MONGO_USER")+
		":"+os.Getenv("MONGO_PASSWORD")+"@"+
		os.Getenv("MONGO_HOST")+
		":"+os.Getenv("MONGO_PORT")+"/?authSource=admin", logger)
	defer client.Disconnect(ctx)
	db := client.Database("tenderdb")
	// logger.Info("Подключено к MongoDB")
	app := fiber.New()
	app.Use(swagger.New(swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.yaml",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}))
	// ctx, cancel := context.WithCancel(context.Background())
	agreementRepo := repository.NewRepository[model.Agreement](db.Collection("agreements"), logger)
	variableRepo := repository.NewRepository[model.Variable](db.Collection("variables"), logger)
	dbp.CreateBase(ctx, variableRepo, logger)
	defer cancel()
	agreementHandler := handler.NewAgreementHandler(agreementRepo, logger)
	app.Get("/agreements/:id", agreementHandler.GetAgreementByID)
	go func() {
		startTasks(logger)
	}()
	logger.Info("Сервер запущен на :8080")
	if err := app.Listen(":8080"); err != nil {
		logger.Fatal("Ошибка при запуске сервера", zap.Error(err))
	}
}
