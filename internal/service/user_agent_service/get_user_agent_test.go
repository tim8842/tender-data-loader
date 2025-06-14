package useragentservice

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/tim8842/tender-data-loader/internal/mocks"
	"github.com/tim8842/tender-data-loader/internal/model"
	"go.uber.org/zap"
)

func TestGetUserAgent(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	tests := []struct {
		name           string
		mockResponse   interface{}
		mockError      error
		expectedResult *model.UserAgentResponse
		expectErr      bool
	}{
		{
			name: "valid response",
			mockResponse: model.UserAgentResponse{
				UserAgent: map[string]interface{}{"agent": "TestAgent"},
				Proxy:     map[string]interface{}{"url": "http://proxy"},
			},
			mockError: nil,
			expectedResult: &model.UserAgentResponse{
				UserAgent: map[string]interface{}{"agent": "TestAgent"},
				Proxy:     map[string]interface{}{"url": "http://proxy"},
			},
			expectErr: false,
		},
		{
			name:           "invalid JSON response",
			mockResponse:   []byte(`invalid json`),
			mockError:      nil,
			expectedResult: nil,
			expectErr:      true,
		},
		{
			name:           "request error",
			mockResponse:   nil,
			mockError:      errors.New("request failed"),
			expectedResult: nil,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReq := new(mocks.MockRequester)

			var respBytes []byte
			switch v := tt.mockResponse.(type) {
			case []byte:
				respBytes = v
			default:
				respBytes, _ = json.Marshal(v)
			}

			mockReq.On("Get", ctx, logger, mock.AnythingOfType("string"), 5*time.Second, mock.Anything).
				Return(respBytes, tt.mockError).Once()

			result, err := GetUserAgent(ctx, "http://test.url", logger, mockReq)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
