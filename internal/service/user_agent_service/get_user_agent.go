package useragentservice

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/util/requests"
	"go.uber.org/zap"
)

func GetUserAgent(ctx context.Context, url string, logger *zap.Logger, requester requests.IRequester) (*model.UserAgentResponse, error) {
	var usStruct model.UserAgentResponse
	res, err := requester.Get(ctx, logger, url, 5*time.Second)
	if err != nil {
		return nil, err
	}
	ok := json.Unmarshal(res, &usStruct)
	if ok != nil {
		logger.Debug("Не может привести тип ответа")
		return nil, ok
	}
	return &usStruct, err
}
