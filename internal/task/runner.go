package task

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/variable"
	"github.com/tim8842/tender-data-loader/pkg"
	"go.uber.org/zap"
)

func StartTasks(
	ctx context.Context, logger *zap.Logger, cfg *config.Config,
	agreeRepo agreement.IAgreementRepo, varRepo variable.IVariableRepo,
	custRepo customer.ICustomerRepo,
) {
	runner := pkg.NewTaskRunner(ctx, logger, 1)

	// Регистрируем задачу
	runner.RegisterTask("back_to_now_agreement", NewBackToNowAgreementTask(cfg, agreeRepo, varRepo, custRepo, true))

	runner.Start()      // Запускаем воркеры
	defer runner.Stop() // Гарантированно остановим

	runner.Enqueue("back_to_now_agreement")
}
