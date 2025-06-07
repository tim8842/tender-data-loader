package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"go.uber.org/zap"
)

// AgreementHandler обрабатывает запросы к договорам
type AgreementHandler struct {
	repo   repository.IMongoRepository
	logger *zap.Logger
}

// NewAgreementHandler создает новый handler
func NewAgreementHandler(
	repo repository.IMongoRepository,
	logger *zap.Logger,
) *AgreementHandler {
	return &AgreementHandler{
		repo:   repo,
		logger: logger,
	}
}

// GetAgreementByID godoc
// @Summary Получить договор по ID
// @Description Возвращает договор по его ID
// @Tags agreements
// @Accept json
// @Produce json
// @Param id path string true "ID договора"
// @Success 200 {object} model.Agreement
// @Failure 404 {object} map[string]string
// @Router /agreements/{id} [get]
func (h *AgreementHandler) GetAgreementByID(c *fiber.Ctx) error {
	id := c.Params("id")

	h.logger.Info("Получение договора", zap.String("id", id))
	agreement, err := h.repo.GetByID(c.Context(), id)
	if err != nil {
		h.logger.Warn("Договор не найден", zap.String("id", id), zap.Error(err))
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "agreement not found",
		})
	}

	return c.JSON(agreement)
}
