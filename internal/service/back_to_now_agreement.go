package service

import (
	"context"
	"time"

	"github.com/tim8842/tender-data-loader/internal/util"
	"go.uber.org/zap"
)

func GetUserAgent(ctx context.Context, url string, logger *zap.Logger) (interface{}, error) {
	res, err := util.GetRequest(ctx, url, 5*time.Second, logger)
	return res, err
}
