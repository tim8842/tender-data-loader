package tasks

import (
	"context"
	"os"
	"time"

	"github.com/tim8842/tender-data-loader/internal/service"
	"github.com/tim8842/tender-data-loader/internal/util"
	"go.uber.org/zap"
)

func BackToNowAgreementTask(ctx context.Context, logger *zap.Logger) error {
	for {
		select {
		case <-ctx.Done():
			logger.Info("BackToNowAgreementTask: Context cancelled, exiting.")
			return ctx.Err()
		default:
			util.ExecuteFunctionPipeline(ctx, []any{service.GetUserAgent}, 3, 5*time.Second, os.Getenv("URL_GET_PROXY"), logger)
			time.Sleep(5 * time.Second)
		}
	}
}
