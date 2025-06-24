package uagent

import (
	"context"
	"encoding/json"
	"time"

	"github.com/tim8842/tender-data-loader/pkg/request"
	"go.uber.org/zap"
)

func GetPage(ctx context.Context, logger *zap.Logger, url string, userAgentReponse *UserAgentResponse, req request.IRequester) (interface{}, error) {
	userAgent, proxyUrl := "", ""
	if tmp, ok := userAgentReponse.UserAgent["agent"].(string); ok {
		userAgent = tmp
	}
	if tmp, ok := userAgentReponse.Proxy["url"].(string); ok {
		proxyUrl = tmp
	}
	return req.Get(ctx, logger, url, 8*time.Second, &request.RequestOptions{
		UserAgent: userAgent,
		ProxyUrl:  proxyUrl,
	})
}

func GetUserAgent(ctx context.Context, url string, logger *zap.Logger, requester request.IRequester) (*UserAgentResponse, error) {
	var usStruct UserAgentResponse
	res, err := requester.Get(ctx, logger, url, 5*time.Second)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(res, &usStruct)
	if err != nil {
		logger.Debug("Не может привести тип ответа")
		return nil, err
	}
	return &usStruct, err
}
