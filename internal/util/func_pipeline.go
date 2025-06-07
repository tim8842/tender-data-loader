package util

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"go.uber.org/zap"
)

func callFunction(fn interface{}, args ...interface{}) (interface{}, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	// Проверяем, что это функция
	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("не является функцией")
	}

	// Проверяем количество возвращаемых значений
	if fnType.NumOut() != 2 {
		return nil, fmt.Errorf("функция должна возвращать 2 значения (any, error)")
	}

	// Проверяем, что второй возвращаемый тип - error
	if !fnType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return nil, fmt.Errorf("второй возвращаемый параметр должен быть error")
	}

	// Подготавливаем аргументы
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	// Вызываем функцию
	results := fnValue.Call(in)

	// Возвращаем результаты
	var err error
	if !results[1].IsNil() {
		err = results[1].Interface().(error)
	}

	return results[0].Interface(), err
}

func FuncWrapper(ctx context.Context, fn any, maxRetries int, delay time.Duration, inner interface{}, logger *zap.Logger) (interface{}, error) {

	var result interface{} = inner
	retries := 0
	for retries <= maxRetries {
		logger.Info("Running function\n")

		res, err := callFunction(fn, ctx, result, logger) // Вызов функции с результатом предыдущей функции

		if err == nil {
			result = res
			logger.Info(truncateRunes(fmt.Sprintf("Fuction complete, res: %v\n", result), 355))

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

// func ExecuteFunctionPipeline(ctx context.Context, functions []any, maxRetries int, delay time.Duration, inner interface{}, logger *zap.Logger) (interface{}, error) {

// 	var result interface{} = inner
// 	for i, fn := range functions {
// 		retries := 0
// 		for retries <= maxRetries {
// 			logger.Info(fmt.Sprintf("Running function %d/%d\n", i+1, len(functions)))

// 			res, err := callFunction(fn, ctx, result, logger) // Вызов функции с результатом предыдущей функции

// 			if err == nil {
// 				result = res
// 				logger.Info(fmt.Sprintf("Fuction complete, res: %v\n", result))

// 				break // Если функция выполнена успешно, переходим к следующей
// 			} else {
// 				logger.Error(fmt.Sprintf("Error in fucntion: %v\n", err))

// 				retries++

// 				if retries <= maxRetries {
// 					logger.Info(fmt.Sprintf("Retry function (%d/%d) for %v...\n", retries, maxRetries, delay))
// 					time.Sleep(delay)
// 				} else {
// 					logger.Info("The maximum number of attempts has been exceeded.")
// 					return nil, fmt.Errorf("failed to execute function after %d retries: %w", maxRetries, err)
// 				}
// 			}
// 		}
// 		if retries > maxRetries {
// 			// Этот блок выполняется, если цикл завершился из-за превышения retries
// 			logger.Info("The maximum number of attempts has been exceeded.")
// 			return nil, errors.New("failed to execute function pipeline") // Возвращаем ошибку
// 		}

// 	}

// 	return result, nil // Возвращаем результат последней успешно выполненной функции и nil для ошибки
// }

// func GetPage(ctx context.Context, res service.UserAgentResponse, logger *zap.Logger) (interface{}, error) {
// 	service.GetPage(ctx)
// }
