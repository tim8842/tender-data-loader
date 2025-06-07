package tasks

import (
	"context"
	"fmt"
	"sync"

	"github.com/tim8842/tender-data-loader/internal/repository"
	"go.uber.org/zap"
)

// type TaskHandler interface {
// 	Process(ctx context.Context, logger *zap.Logger) error
// }

// // TaskFunc - тип для функции, реализующей обработку задачи
// type TaskFunc func(ctx context.Context, logger *zap.Logger) error

// // Process реализует интерфейс TaskHandler для TaskFunc
// func (f TaskFunc) Process(ctx context.Context, logger *zap.Logger) error {
// 	return f(ctx, logger)
// }

// func Worker(ctx context.Context, id int, tasks <-chan string, handlers map[string]TaskHandler, wg *sync.WaitGroup, logger *zap.Logger) {
// 	defer wg.Done()

// 	logger.Info(fmt.Sprintf("Worker %d: Starting...", id))
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			logger.Info(fmt.Sprintf("Worker %d: Context cancelled, exiting.", id))
// 			return
// 		case taskName, ok := <-tasks:
// 			if !ok {
// 				logger.Info(fmt.Sprintf("Worker %d: Task channel closed, exiting.", id))
// 				return
// 			}

// 			handler, ok := handlers[taskName]
// 			if !ok {
// 				logger.Warn(fmt.Sprintf("Worker %d: No handler registered for task %s", id, taskName))
// 				continue // Переходим к следующей задаче
// 			}

// 			logger.Info(fmt.Sprintf("Worker %d: Processing task %s", id, taskName))

// 			err := handler.Process(ctx, logger)
// 			if err != nil {
// 				logger.Error(fmt.Sprintf("Worker %d: Error processing task %s: %v", id, taskName, err))
// 			}

// 			logger.Info(fmt.Sprintf("Worker %d: Finished task %s", id, taskName))
// 		}
// 	}
// }

type TaskHandlerWithRepo interface {
	Process(ctx context.Context, logger *zap.Logger, repo repository.IMongoRepository) error
}

// TaskFunc - тип для функции, реализующей обработку задачи
type TaskFuncWithRepo func(ctx context.Context, logger *zap.Logger, repo repository.IMongoRepository) error

// Process реализует интерфейс TaskHandler для TaskFunc
func (f TaskFuncWithRepo) Process(ctx context.Context, logger *zap.Logger, repo repository.IMongoRepository) error {
	return f(ctx, logger, repo)
}

func WorkerWithRepo(ctx context.Context, id int, tasks <-chan string, handlers map[string]TaskHandlerWithRepo, wg *sync.WaitGroup, logger *zap.Logger, repo repository.IMongoRepository) {
	defer wg.Done()

	logger.Info(fmt.Sprintf("Worker %d: Starting...", id))
	for {
		select {
		case <-ctx.Done():
			logger.Info(fmt.Sprintf("Worker %d: Context cancelled, exiting.", id))
			return
		case taskName, ok := <-tasks:
			if !ok {
				logger.Info(fmt.Sprintf("Worker %d: Task channel closed, exiting.", id))
				return
			}

			handler, ok := handlers[taskName]
			if !ok {
				logger.Warn(fmt.Sprintf("Worker %d: No handler registered for task %s", id, taskName))
				continue // Переходим к следующей задаче
			}

			logger.Info(fmt.Sprintf("Worker %d: Processing task %s", id, taskName))

			err := handler.Process(ctx, logger, repo)
			if err != nil {
				logger.Error(fmt.Sprintf("Worker %d: Error processing task %s: %v", id, taskName, err))
			}

			logger.Info(fmt.Sprintf("Worker %d: Finished task %s", id, taskName))
		}
	}
}

type TaskHandler interface {
	Process(ctx context.Context, logger *zap.Logger, data map[string]any) error
}

// TaskFunc - тип для функции, реализующей обработку задачи
type TaskFunc func(ctx context.Context, logger *zap.Logger, data map[string]any) error

// Process реализует интерфейс TaskHandler для TaskFunc
func (f TaskFunc) Process(ctx context.Context, logger *zap.Logger, data map[string]any) error {
	return f(ctx, logger, data)
}

func Worker(ctx context.Context, id int, tasks <-chan string, handlers map[string]TaskHandler, wg *sync.WaitGroup, logger *zap.Logger, data map[string]any) {
	defer wg.Done()

	logger.Info(fmt.Sprintf("Worker %d: Starting...", id))
	for {
		select {
		case <-ctx.Done():
			logger.Info(fmt.Sprintf("Worker %d: Context cancelled, exiting.", id))
			return
		case taskName, ok := <-tasks:
			if !ok {
				logger.Info(fmt.Sprintf("Worker %d: Task channel closed, exiting.", id))
				return
			}

			handler, ok := handlers[taskName]
			if !ok {
				logger.Warn(fmt.Sprintf("Worker %d: No handler registered for task %s", id, taskName))
				continue // Переходим к следующей задаче
			}

			logger.Info(fmt.Sprintf("Worker %d: Processing task %s", id, taskName))

			err := handler.Process(ctx, logger, data)
			if err != nil {
				logger.Error(fmt.Sprintf("Worker %d: Error processing task %s: %v", id, taskName, err))
			}

			logger.Info(fmt.Sprintf("Worker %d: Finished task %s", id, taskName))
		}
	}
}
