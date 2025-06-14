package setup

import (
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/tim8842/tender-data-loader/internal/handler"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"go.uber.org/zap"
)

func SetupFiberApp(repositories *repository.Repositories, logger *zap.Logger) *fiber.App {
	app := fiber.New()

	app.Use(swagger.New(swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.yaml",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}))

	agreementHandler := handler.NewAgreementHandler(repositories, logger)
	app.Get("/agreements/:id", agreementHandler.GetAgreementByID)

	return app
}
