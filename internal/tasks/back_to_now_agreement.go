package tasks

import (
	"context"
	"time"

	"go.uber.org/zap"
)

func BackToNowAgreementTask(ctx context.Context, logger *zap.Logger) error {
	for {
		select {
		case <-ctx.Done():
			logger.Info("BackToNowAgreementTask: Context cancelled, exiting.")
			return ctx.Err()
		default:
			logger.Info("BackToNowAgreementTask: DSADASDASDSDAS")
			time.Sleep(5 * time.Second)
		}
	}
}
