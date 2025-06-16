package requests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestRequester_Get(t *testing.T) {
	type testCase struct {
		name        string
		handler     http.HandlerFunc
		timeout     time.Duration
		expectError bool
		opts        *RequestOptions
	}

	tests := []testCase{
		{
			name: "успешный ответ 200",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "ok"}`))
			},
			timeout:     2 * time.Second,
			expectError: false,
		},
		{
			name: "сервер возвращает 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			timeout:     2 * time.Second,
			expectError: true,
		},
		{
			name: "таймаут превышен",
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
			timeout:     500 * time.Millisecond,
			expectError: true,
		},
		{
			name: "установка User-Agent",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("User-Agent") != "MyTestAgent" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.WriteHeader(http.StatusOK)
			},
			timeout: 1 * time.Second,
			opts: &RequestOptions{
				UserAgent: "MyTestAgent",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		tt := tt // захват переменной
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(tt.handler)
			defer server.Close()

			r := &Requester{}
			logger := zaptest.NewLogger(t)
			resp, err := r.Get(context.Background(), logger, server.URL, tt.timeout, tt.opts)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestRequester_Get_InvalidProxy(t *testing.T) {
	r := &Requester{}
	logger := zaptest.NewLogger(t)

	_, err := r.Get(context.Background(), logger, "http://example.com", 1*time.Second, &RequestOptions{
		ProxyUrl: "http://[::1]:namedport", // некорректный URL
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ошибка при парсинге URL прокси")
}
