package backtonowagreementservice

import (
	"context"
	"time"

	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/util/requests"
	"go.uber.org/zap"
)

func GetPage(ctx context.Context, logger *zap.Logger, url string, userAgentReponse *model.UserAgentResponse, req requests.IRequester) (interface{}, error) {
	userAgent, proxyUrl := "", ""
	if tmp, ok := userAgentReponse.UserAgent["agent"].(string); ok {
		userAgent = tmp
	}
	if tmp, ok := userAgentReponse.Proxy["url"].(string); ok {
		proxyUrl = tmp
	}
	return req.Get(ctx, logger, url, 5*time.Second, &requests.RequestOptions{
		UserAgent: userAgent,
		ProxyUrl:  proxyUrl,
	})
}
