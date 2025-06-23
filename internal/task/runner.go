package task

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/contract"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/supplier"
	"github.com/tim8842/tender-data-loader/internal/variable"
	"github.com/tim8842/tender-data-loader/pkg"
	"go.uber.org/zap"
)

func StartTasks(
	ctx context.Context, logger *zap.Logger, cfg *config.Config,
	agreeRepo agreement.IAgreementRepo, varRepo variable.IVariableRepo,
	custRepo customer.ICustomerRepo,
	suppRepo supplier.ISupplierRepo,
	contract contract.IContractRepo,
) {
	runner := pkg.NewTaskRunner(ctx, logger, 7)

	// Регистрируем задачу
	// runner.RegisterTask("back_to_now_agreement", NewBackToNowAgreementTask(cfg, agreeRepo, varRepo, custRepo, true))
	runner.RegisterTask("back_to_now_contract40000", NewBackToNowContractTask(cfg, contract, varRepo, custRepo, suppRepo, "back_to_now_contract40000", false))
	// runner.RegisterTask("back_to_now_contract100000", NewBackToNowContractTask(cfg, contract, varRepo, custRepo, suppRepo, "back_to_now_contract100000", true))
	// runner.RegisterTask("back_to_now_contract300000", NewBackToNowContractTask(cfg, contract, varRepo, custRepo, suppRepo, "back_to_now_contract300000", true))
	// runner.RegisterTask("back_to_now_contract600000", NewBackToNowContractTask(cfg, contract, varRepo, custRepo, suppRepo, "back_to_now_contract600000", true))
	// runner.RegisterTask("back_to_now_contract10000000", NewBackToNowContractTask(cfg, contract, varRepo, custRepo, suppRepo, "back_to_now_contract10000000", true))
	// runner.RegisterTask("back_to_now_contract999999999999", NewBackToNowContractTask(cfg, contract, varRepo, custRepo, suppRepo, "back_to_now_contract999999999999", true))

	runner.Start()      // Запускаем воркеры
	defer runner.Stop() // Гарантированно остановим

	// runner.Enqueue("back_to_now_agreement")
	runner.Enqueue("back_to_now_contract40000")
	// runner.Enqueue("back_to_now_contract100000")
	// runner.Enqueue("back_to_now_contract300000")
	// runner.Enqueue("back_to_now_contract600000")
	// runner.Enqueue("back_to_now_contract10000000")
	// runner.Enqueue("back_to_now_contract999999999999")
}
