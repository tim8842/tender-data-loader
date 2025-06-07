package requests

import (
	"context"
	"fmt"
	"io"
	"net/http"
	urlPackage "net/url"
	"time"

	"go.uber.org/zap"
)

// RequestOptions определяет опции для HTTP-запроса.
type RequestOptions struct {
	UserAgent string // Пользовательский агент (User-Agent)\
	ProxyUrl  string
}

// getRequest выполняет HTTP GET-запрос к указанному URL и возвращает JSON-данные.
//
//	Параметры:
//	 - ctx:      Контекст для управления запросом (таймаут, отмена).
//	 - url:      URL для GET-запроса.
//	 - timeout:  Таймаут для запроса.
//	 - logger:   Указатель на логгер (для логирования).
//	 - options:  Необязательные опции (RequestOptions).
//	Возвращает:
//	 - интерфейс{} (то есть map[string]interface{} или []interface{} в зависимости от JSON)
//	 - ошибку (если произошла).
func GetRequest(ctx context.Context, url string, timeout time.Duration, logger *zap.Logger, options ...RequestOptions) ([]byte, error) {
	// 1. Создаем контекст с таймаутом (если не передан)
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка при создании запроса: %v", err))
		return nil, fmt.Errorf("ошибка при создании запроса: %w", err)
	}
	client := &http.Client{}
	// 3. Устанавливаем опции (User-Agent)
	if len(options) > 0 {
		opts := options[0] // Берем первую опцию

		// Устанавливаем User-Agent
		if opts.UserAgent != "" {
			req.Header.Set("User-Agent", opts.UserAgent)
		}

		// Настройка прокси, если указан
		if opts.ProxyUrl != "" {
			proxyURL, err := urlPackage.Parse(opts.ProxyUrl)
			if err != nil {
				logger.Error(fmt.Sprintf("Ошибка при парсинге URL прокси: %v", err))
				return nil, fmt.Errorf("ошибка при парсинге URL прокси: %w", err)
			}

			transport := &http.Transport{
				Proxy: http.ProxyURL(proxyURL), // Используем URL прокси
			}
			client = &http.Client{ //Переопределяем Client при наличии прокси
				Transport: transport,
			}
		}
	}

	// 4. Выполняем запрос

	logger.Info(fmt.Sprintf("Отправка запроса GET: %s", url))
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка при выполнении запроса: %v", err))
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer resp.Body.Close() //  Обязательно закрываем Body

	// 5. Обрабатываем статус ответа
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn(fmt.Sprintf("Неверный статус ответа: %d", resp.StatusCode))
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	// 6. Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Ошибка при чтении тела ответа: %v", err))
		return nil, fmt.Errorf("ошибка при чтении тела ответа: %w", err)
	}
	return body, nil
}
