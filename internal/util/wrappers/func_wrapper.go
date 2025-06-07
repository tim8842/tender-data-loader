package wrappers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tim8842/tender-data-loader/internal/util/base"
	"go.uber.org/zap"
)

type FuncInWrapp interface {
	Process(ctx context.Context, logger *zap.Logger) (any, error)
}

func FuncWrapper(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn FuncInWrapp) (any, error) {
	// Функция обертка для того, чтобы функция пыталась выполнится еще несколько раз, если не получилось
	var result any = nil
	retries := 0
	for retries <= maxRetries {
		res, err := fn.Process(ctx, logger)

		if err == nil {
			result = res
			logger.Info(base.TruncateRunes(fmt.Sprintf("Fuction complete, res: %v\n", result), 355))

			break // Если функция выполнена успешно, переходим к следующей
		} else {

			retries++

			if retries <= maxRetries {
				logger.Info(fmt.Sprintf("Retry function (%d/%d) for %v...\n", retries, maxRetries, delay))
				time.Sleep(delay)
			} else {
				logger.Info("The maximum number of attempts has been exceeded.")
				return nil, fmt.Errorf("failed to execute function after %d retries: %w", maxRetries, err)
			}
		}
	}
	if retries > maxRetries {
		// Этот блок выполняется, если цикл завершился из-за превышения retries
		logger.Info("The maximum number of attempts has been exceeded.")
		return nil, errors.New("failed to execute function pipeline") // Возвращаем ошибку
	}

	return result, nil // Возвращаем результат последней успешно выполненной функции и nil для ошибки
}
