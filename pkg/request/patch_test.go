package request_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tim8842/tender-data-loader/pkg/request"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel) // без вывода
	logger, _ := cfg.Build()
	return logger
}

func TestPatch_TableDriven(t *testing.T) {
	logger := newTestLogger()
	ctx := context.Background()

	tests := []struct {
		name        string
		handlerFunc http.HandlerFunc
		url         string
		body        []byte
		headers     map[string]string
		timeout     time.Duration
		wantErr     bool
		errContains string
	}{
		{
			name: "success 200",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected method PATCH, got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`ok`))
			},
			body:    []byte(`data`),
			headers: map[string]string{"X-Test": "val"},
			timeout: 5 * time.Second,
			wantErr: false,
		},
		{
			name: "bad status 400",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "bad", http.StatusBadRequest)
			},
			timeout:     5 * time.Second,
			wantErr:     true,
			errContains: "неверный статус ответа",
		},
		{
			name: "timeout",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
			timeout:     1 * time.Second,
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
		{
			name:        "invalid url",
			url:         ":",
			wantErr:     true,
			errContains: "ошибка создания запроса",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.url
			if url == "" && tt.handlerFunc != nil {
				ts := httptest.NewServer(tt.handlerFunc)
				defer ts.Close()
				url = ts.URL
			}

			resp, err := request.Patch(ctx, logger, url, tt.body, tt.headers, tt.timeout)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error status: got err=%v, wantErr=%v", err, tt.wantErr)
			}
			if err != nil && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
				t.Errorf("error message %q does not contain expected %q", err.Error(), tt.errContains)
			}

			if err == nil && resp == nil {
				t.Errorf("expected non-nil response when no error")
			}
		})
	}
}

func TestPatchJSON(t *testing.T) {
	logger := newTestLogger()
	ctx := context.Background()

	type Payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name        string
		handlerFunc http.HandlerFunc
		payload     interface{}
		timeout     time.Duration
		wantErr     bool
		errContains string
	}{
		{
			name: "success 200",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected method PATCH, got %s", r.Method)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", ct)
				}

				var p Payload
				err := json.NewDecoder(r.Body).Decode(&p)
				if err != nil {
					t.Errorf("failed decode json body: %v", err)
				}
				if p.Name != "Alice" || p.Age != 30 {
					t.Errorf("unexpected payload: %+v", p)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`ok`))
			},
			payload: Payload{Name: "Alice", Age: 30},
			timeout: 5 * time.Second,
			wantErr: false,
		},
		{
			name:        "json marshal error",
			payload:     func() {}, // функция не маршалится в JSON, ошибка
			timeout:     5 * time.Second,
			wantErr:     true,
			errContains: "json: unsupported type",
		},
		{
			name: "timeout",
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(2 * time.Second)
				w.WriteHeader(http.StatusOK)
			},
			payload:     Payload{Name: "Alice", Age: 30},
			timeout:     1 * time.Second,
			wantErr:     true,
			errContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := ""
			if tt.handlerFunc != nil {
				ts := httptest.NewServer(tt.handlerFunc)
				defer ts.Close()
				url = ts.URL
			}

			resp, err := request.PatchJSON(ctx, logger, url, tt.payload, tt.timeout)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got err: %v", tt.wantErr, err)
			}
			if err != nil && tt.errContains != "" && !contains(err.Error(), tt.errContains) {
				t.Errorf("error message %q does not contain expected %q", err.Error(), tt.errContains)
			}
			if err == nil && resp == nil {
				t.Errorf("expected non-nil response when no error")
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && (string(s[0:len(substr)]) == substr || contains(s[1:], substr)))
}
