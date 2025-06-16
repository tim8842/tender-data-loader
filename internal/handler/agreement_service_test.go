package handler

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/mocks"
	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Тестируем успешную загрузку всех данных по id в хэндлере
func TestAgreementHandler_GetAgreementByID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockReturn     *model.Agreement
		mockError      error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful get",
			id:   "123",
			mockReturn: &model.Agreement{
				ID:     "123",
				Number: "Test Agreement",
			},
			mockError:      nil,
			expectedStatus: fiber.StatusOK,
			expectedBody:   `{"customer_id":"", "execution_end":"0001-01-01T00:00:00Z", "execution_start":"0001-01-01T00:00:00Z", "id":"123", "notice_id":"", "number":"Test Agreement", "pdif":"", "price":0, "published_at":"0001-01-01T00:00:00Z", "purchase_method":"", "services": [], "signed_at":"0001-01-01T00:00:00Z", "status":"", "subject":"", "updated_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name:           "not found",
			id:             "456",
			mockReturn:     nil,
			mockError:      mongo.ErrNoDocuments,
			expectedStatus: fiber.StatusNotFound,
			expectedBody:   `{"error":"agreement not found"}`,
		},
		{
			name:           "internal error",
			id:             "789",
			mockReturn:     nil,
			mockError:      errors.New("internal error"),
			expectedStatus: fiber.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			logger, _ := zap.NewProduction()
			defer logger.Sync()
			mockRepo := new(mocks.MockGenericRepository[*model.Agreement])
			mockRepo.On("GetByID", mock.Anything, tt.id).Return(tt.mockReturn, tt.mockError)
			repos := &repository.Repositories{
				AgreementRepo: mockRepo,
			}
			h := NewAgreementHandler(repos, logger)
			app.Get("/agreements/:id", h.GetAgreementByID)
			req := httptest.NewRequest("GET", "/agreements/"+tt.id, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			body := make([]byte, resp.ContentLength)
			resp.Body.Read(body)
			assert.JSONEq(t, tt.expectedBody, string(body))
			mockRepo.AssertCalled(t, "GetByID", mock.Anything, tt.id)
		})
	}
}
