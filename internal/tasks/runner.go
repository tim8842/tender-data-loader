package tasks

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/repository"
	"go.uber.org/zap"
)

func StartTasks(ctx context.Context, logger *zap.Logger, varRepo repository.IMongoRepository) {
	runner := NewTaskRunner(ctx, logger, 1)

	// Регистрируем задачу
	runner.RegisterTask("back_to_now_agreement", NewBackToNowAgreementTask(varRepo))

	runner.Start()      // Запускаем воркеры
	defer runner.Stop() // Гарантированно остановим

	runner.Enqueue("back_to_now_agreement")
}
