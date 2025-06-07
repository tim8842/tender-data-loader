package tasks

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// TaskHandler интерфейс для задачи с инъекцией зависимостей через конструктор
type TaskHandler interface {
	Process(ctx context.Context, logger *zap.Logger) error
}

// TaskRunner управляет запуском воркеров и тасок
type TaskRunner struct {
	logger      *zap.Logger
	ctx         context.Context
	workerCount int
	tasksChan   chan string
	handlers    map[string]TaskHandler
	wg          sync.WaitGroup
}

func NewTaskRunner(ctx context.Context, logger *zap.Logger, workerCount int) *TaskRunner {
	return &TaskRunner{
		ctx:         ctx,
		logger:      logger,
		workerCount: workerCount,
		tasksChan:   make(chan string),
		handlers:    make(map[string]TaskHandler),
	}
}

func (r *TaskRunner) RegisterTask(name string, handler TaskHandler) {
	r.handlers[name] = handler
}

func (r *TaskRunner) Start() {
	r.logger.Info("Starting task runner", zap.Int("workers", r.workerCount))
	r.wg.Add(r.workerCount)
	for i := 0; i < r.workerCount; i++ {
		go r.worker(i + 1)
	}
}

func (r *TaskRunner) Stop() {
	r.logger.Info("Stopping task runner")
	close(r.tasksChan)
	r.wg.Wait()
	r.logger.Info("All workers stopped")
}

func (r *TaskRunner) Enqueue(taskName string) {
	select {
	case r.tasksChan <- taskName:
		r.logger.Info("Task enqueued", zap.String("task", taskName))
	case <-r.ctx.Done():
		r.logger.Warn("Failed to enqueue task, context done", zap.String("task", taskName))
	}
}

func (r *TaskRunner) worker(id int) {
	defer r.wg.Done()
	r.logger.Info("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-r.ctx.Done():
			r.logger.Info("Worker context cancelled", zap.Int("worker_id", id))
			return
		case taskName, ok := <-r.tasksChan:
			if !ok {
				r.logger.Info("Tasks channel closed", zap.Int("worker_id", id))
				return
			}

			handler, ok := r.handlers[taskName]
			if !ok {
				r.logger.Warn("No handler found for task", zap.Int("worker_id", id), zap.String("task", taskName))
				continue
			}

			r.logger.Info("Processing task", zap.Int("worker_id", id), zap.String("task", taskName))
			err := handler.Process(r.ctx, r.logger)
			if err != nil {
				r.logger.Error("Error processing task", zap.Int("worker_id", id), zap.String("task", taskName), zap.Error(err))
			}
			r.logger.Info("Finished task", zap.Int("worker_id", id), zap.String("task", taskName))
		}
	}
}

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

// type TaskHandlerWithRepo interface {
// 	Process(ctx context.Context, logger *zap.Logger, repo repository.IMongoRepository) error
// }

// // TaskFunc - тип для функции, реализующей обработку задачи
// type TaskFuncWithRepo func(ctx context.Context, logger *zap.Logger, repo repository.IMongoRepository) error

// // Process реализует интерфейс TaskHandler для TaskFunc
// func (f TaskFuncWithRepo) Process(ctx context.Context, logger *zap.Logger, repo repository.IMongoRepository) error {
// 	return f(ctx, logger, repo)
// }

// func WorkerWithRepo(ctx context.Context, id int, tasks <-chan string, handlers map[string]TaskHandlerWithRepo, wg *sync.WaitGroup, logger *zap.Logger, repo repository.IMongoRepository) {
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

// 			err := handler.Process(ctx, logger, repo)
// 			if err != nil {
// 				logger.Error(fmt.Sprintf("Worker %d: Error processing task %s: %v", id, taskName, err))
// 			}

// 			logger.Info(fmt.Sprintf("Worker %d: Finished task %s", id, taskName))
// 		}
// 	}
// }

// type TaskHandler interface {
// 	Process(ctx context.Context, logger *zap.Logger, data map[string]any) error
// }

// // TaskFunc - тип для функции, реализующей обработку задачи
// type TaskFunc func(ctx context.Context, logger *zap.Logger, data map[string]any) error

// // Process реализует интерфейс TaskHandler для TaskFunc
// func (f TaskFunc) Process(ctx context.Context, logger *zap.Logger, data map[string]any) error {
// 	return f(ctx, logger, data)
// }

// func Worker(ctx context.Context, id int, tasks <-chan string, handlers map[string]TaskHandler, wg *sync.WaitGroup, logger *zap.Logger, data map[string]any) {
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

// 			err := handler.Process(ctx, logger, data)
// 			if err != nil {
// 				logger.Error(fmt.Sprintf("Worker %d: Error processing task %s: %v", id, taskName, err))
// 			}

// 			logger.Info(fmt.Sprintf("Worker %d: Finished task %s", id, taskName))
// 		}
// 	}
// }
