package request

import (
	"compress/flate"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	urlPackage "net/url"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/txthinking/socks5"
	"go.uber.org/zap"
)

type RequestOptions struct {
	UserAgent string // Пользовательский агент (User-Agent)\
	ProxyUrl  string
}

type IRequester interface {
	Get(ctx context.Context, logger *zap.Logger, url string, timeout time.Duration, opts ...*RequestOptions) ([]byte, error)
}
type Requester struct{}

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

func (d *Requester) Get(ctx context.Context, logger *zap.Logger, url string, timeout time.Duration, options ...*RequestOptions) ([]byte, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Error("Ошибка при создании запроса", zap.Error(err))
		return nil, fmt.Errorf("ошибка при создании запроса: %w", err)
	}

	// Ставим общие заголовки всегда
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{Timeout: timeout}

	if len(options) > 0 && options[0] != nil {
		opts := options[0]

		if opts.UserAgent != "" {
			req.Header.Set("User-Agent", opts.UserAgent)
		}

		if opts.ProxyUrl != "" {
			proxyURL, err := urlPackage.Parse(opts.ProxyUrl)
			if err != nil {
				logger.Error("Ошибка при парсинге URL прокси", zap.Error(err))
				return nil, fmt.Errorf("ошибка при парсинге URL прокси: %w", err)
			}

			switch proxyURL.Scheme {
			case "http", "https":
				transport := &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
				client.Transport = transport

			case "socks5", "socks4":
				address := proxyURL.Host
				if !strings.Contains(address, ":") {
					address += ":1080"
				}

				username := ""
				password := ""
				if proxyURL.User != nil {
					username = proxyURL.User.Username()
					password, _ = proxyURL.User.Password()
				}

				clientSocks5, err := socks5.NewClient(address, username, password, int(timeout.Seconds()), 0)
				if err != nil {
					logger.Error("Ошибка при создании SOCKS5 клиента", zap.Error(err))
					return nil, fmt.Errorf("ошибка при создании SOCKS5 клиента: %w", err)
				}

				dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
					host, port, _ := net.SplitHostPort(addr)
					ips, err := net.LookupIP(host)
					if err == nil && len(ips) > 0 {
						realAddr := net.JoinHostPort(ips[0].String(), port)
						return clientSocks5.Dial(network, realAddr)
					}
					logger.Warn("DNS lookup failed, пытаемся напрямую", zap.String("host", host), zap.Error(err))
					// fallback: пробуем напрямую, пусть прокси резолвит
					return clientSocks5.Dial(network, addr)
				}

				transport := &http.Transport{
					DialContext: dialContext,
				}
				client.Transport = transport

			default:
				logger.Warn("Неподдерживаемый тип прокси", zap.String("scheme", proxyURL.Scheme))
			}
		}
	}

	logger.Debug("Отправка запроса GET", zap.String("url", url))
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Ошибка при выполнении запроса", zap.Error(err))
		return nil, fmt.Errorf("ошибка при выполнении запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn("Неверный статус ответа", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			logger.Error("Ошибка распаковки gzip", zap.Error(err))
			return nil, fmt.Errorf("ошибка распаковки gzip: %w", err)
		}
		defer reader.Close()
	case "br":
		reader = io.NopCloser(brotli.NewReader(resp.Body))
	case "deflate":
		reader = flate.NewReader(resp.Body)
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		logger.Error("Ошибка при чтении тела ответа", zap.Error(err))
		return nil, fmt.Errorf("ошибка при чтении тела ответа: %w", err)
	}

	return body, nil
}
