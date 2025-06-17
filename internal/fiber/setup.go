package fiber

import (
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/tim8842/tender-data-loader/internal/agreement"
	"go.uber.org/zap"
)

func SetupFiberApp(
	logger *zap.Logger, agreePero agreement.IAgreementRepo,
) *fiber.App {
	app := fiber.New()

	app.Use(swagger.New(swagger.Config{
		BasePath: "/",
		FilePath: "./docs/swagger.yaml",
		Path:     "swagger",
		Title:    "Swagger API Docs",
	}))

	agreementHandler := agreement.NewAgreementHandler(logger, agreePero)
	app.Get("/agreements/:id", agreementHandler.GetAgreementByID)

	return app
}
