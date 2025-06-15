package subtasks

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/model"
	backtonowagreementservice "github.com/tim8842/tender-data-loader/internal/service/back_to_now_agreement_service"
	useragentservice "github.com/tim8842/tender-data-loader/internal/service/user_agent_service"
	"github.com/tim8842/tender-data-loader/internal/util/requests"
	"go.uber.org/zap"
)

// Обычный гет запрос

type GetRequest struct {
	url string
}

func NewGetRequest(url string) *GetRequest {
	return &GetRequest{url: url}
}

func (t GetRequest) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := useragentservice.GetUserAgent(ctx, t.url, logger, &requests.Requester{})
	if ok != nil {
		return nil, ok
	}
	return data, ok
}

// Гет запрос с проски и UserAgent

type GetPage struct {
	url               string
	userAgentResponse *model.UserAgentResponse
}

func NewGetPage(url string, userAgentResponse *model.UserAgentResponse) *GetPage {
	return &GetPage{url: url, userAgentResponse: userAgentResponse}
}

func (t GetPage) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	req := &requests.Requester{}
	data, err := backtonowagreementservice.GetPage(ctx, logger, t.url, t.userAgentResponse, req)
	return data, err
}
