package backtonowagreementservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/mocks"
	"github.com/tim8842/tender-data-loader/internal/model"
	"go.uber.org/zap"
)

// Мок реализация интерфейса IRequester

func TestGetPage(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name          string
		userAgentResp *model.UserAgentResponse
		url           string
		mockReturn    []byte
		mockErr       error
		expectErr     bool
	}{
		{
			name: "successful request",
			userAgentResp: &model.UserAgentResponse{
				UserAgent: map[string]interface{}{"agent": "Mozilla/5.0"},
				Proxy:     map[string]interface{}{"url": "http://localhost:8080"},
			},
			url:        "https://example.com",
			mockReturn: []byte(`{"status":"ok"}`),
			mockErr:    nil,
			expectErr:  false,
		},
		{
			name: "request fails",
			userAgentResp: &model.UserAgentResponse{
				UserAgent: map[string]interface{}{"agent": "BadBot"},
				Proxy:     map[string]interface{}{"url": "http://badproxy"},
			},
			url:        "https://fail.com",
			mockReturn: nil,
			mockErr:    errors.New("network error"),
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReq := new(mocks.MockRequester)

			mockReq.
				On("Get", ctx, mock.Anything, tt.url, 5*time.Second, mock.Anything).
				Return(tt.mockReturn, tt.mockErr).
				Once()

			res, err := GetPage(ctx, logger, tt.url, tt.userAgentResp, mockReq)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockReturn, res)
			}

			mockReq.AssertExpectations(t)
		})
	}
}
