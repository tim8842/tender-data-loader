package uagent_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	inmock "github.com/tim8842/tender-data-loader/internal/mock"
	"github.com/tim8842/tender-data-loader/internal/uagent"
	"go.uber.org/zap"
)

func TestGetPage(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name          string
		userAgentResp *uagent.UserAgentResponse
		url           string
		mockReturn    []byte
		mockErr       error
		expectErr     bool
	}{
		{
			name: "successful request",
			userAgentResp: &uagent.UserAgentResponse{
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
			userAgentResp: &uagent.UserAgentResponse{
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
			mockReq := new(inmock.MockRequester)

			mockReq.
				On("Get", ctx, mock.Anything, tt.url, 8*time.Second, mock.Anything).
				Return(tt.mockReturn, tt.mockErr).
				Once()

			res, err := uagent.GetPage(ctx, logger, tt.url, tt.userAgentResp, mockReq)
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

func TestGetUserAgent(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewNop()

	tests := []struct {
		name           string
		mockResponse   interface{}
		mockError      error
		expectedResult *uagent.UserAgentResponse
		expectErr      bool
	}{
		{
			name: "valid response",
			mockResponse: uagent.UserAgentResponse{
				UserAgent: map[string]interface{}{"agent": "TestAgent"},
				Proxy:     map[string]interface{}{"url": "http://proxy"},
			},
			mockError: nil,
			expectedResult: &uagent.UserAgentResponse{
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
			mockReq := new(inmock.MockRequester)

			var respBytes []byte
			switch v := tt.mockResponse.(type) {
			case []byte:
				respBytes = v
			default:
				respBytes, _ = json.Marshal(v)
			}

			mockReq.On("Get", ctx, logger, mock.AnythingOfType("string"), 5*time.Second, mock.Anything).
				Return(respBytes, tt.mockError).Once()

			result, err := uagent.GetUserAgent(ctx, "http://test.url", logger, mockReq)

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
