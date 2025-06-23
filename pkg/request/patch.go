package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Patch делает HTTP PATCH запрос на указанный url с переданным телом и заголовками.
// Если timeout > 0, создаётся контекст с таймаутом.
func Patch(ctx context.Context, logger *zap.Logger, url string, body []byte, headers map[string]string, timeout time.Duration) ([]byte, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Если не указан Content-Type, ставим по умолчанию application/json
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем код ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("получен неверный статус ответа: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}
	fmt.Println(string(body))
	logger.Error("ASDASDASDASDS " + string(data))
	return data, nil
}

func PatchJSON(ctx context.Context, logger *zap.Logger, url string, payload interface{}, timeout time.Duration) ([]byte, error) {
	fmt.Println(payload)
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return Patch(ctx, logger, url, body, map[string]string{
		"Content-Type": "application/json",
	}, timeout)
}
