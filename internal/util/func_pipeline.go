package util

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func executeFunctionPipeline(functions []func(interface{}) (interface{}, error), maxRetries int, delay time.Duration, inner interface{}, logger *zap.Logger) (interface{}, error) {
	var result interface{} = inner
	for i, fn := range functions {
		retries := 0
		for retries <= maxRetries {
			logger.Info(fmt.Sprintf("Running function %d/%d\n", i+1, len(functions)))

			res, err := fn(result) // Вызов функции с результатом предыдущей функции

			if err == nil {
				result = res
				logger.Info(fmt.Sprintf("Fuction complete, res: %v\n", result))

				break // Если функция выполнена успешно, переходим к следующей
			} else {
				logger.Error(fmt.Sprintf("Error in fucntion: %v\n", err))

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

	}

	return result, nil // Возвращаем результат последней успешно выполненной функции и nil для ошибки
}
