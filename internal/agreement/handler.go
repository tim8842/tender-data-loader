package agreement

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// AgreementHandler обрабатывает запросы к договорам
type AgreementHandler struct {
	logger    *zap.Logger
	agreeRepo IAgreementRepo
}

// NewAgreementHandler создает новый handler
func NewAgreementHandler(
	logger *zap.Logger,
	agreeRepo IAgreementRepo,
) *AgreementHandler {
	return &AgreementHandler{
		logger:    logger,
		agreeRepo: agreeRepo,
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
	agreement, err := h.agreeRepo.GetByID(c.Context(), id)
	if agreement != nil {
		agreement.Services = []*AgreementService{}
	}
	if err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) || strings.Contains(err.Error(), "not found") {
			// Ошибка "не найдено" - возвращаем 404
			h.logger.Warn("Договор не найден", zap.String("id", id), zap.Error(err))
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "agreement not found",
			})
		}

		// Все остальные ошибки - возвращаем 500
		h.logger.Error("Ошибка при получении договора", zap.String("id", id), zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(agreement)
}
