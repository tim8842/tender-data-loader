package uagent

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/uagent"
	"github.com/tim8842/tender-data-loader/pkg/request"
	"go.uber.org/zap"
)

type GetRequest struct {
	url string
}

func NewGetRequest(url string) *GetRequest {
	return &GetRequest{url: url}
}

func (t GetRequest) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := uagent.GetUserAgent(ctx, t.url, logger, &request.Requester{})
	if ok != nil {
		return nil, ok
	}
	return data, ok
}

// Гет запрос с проски и UserAgent

type GetPage struct {
	url               string
	userAgentResponse *uagent.UserAgentResponse
}

func NewGetPage(url string, userAgentResponse *uagent.UserAgentResponse) *GetPage {
	return &GetPage{url: url, userAgentResponse: userAgentResponse}
}

func (t GetPage) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	req := &request.Requester{}
	data, err := uagent.GetPage(ctx, logger, t.url, t.userAgentResponse, req)
	return data, err
}
