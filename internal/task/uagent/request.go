package uagent

import (
	"context"
	"time"

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
	data, err := uagent.GetUserAgent(ctx, t.url, logger, &request.Requester{})
	if err != nil {
		return nil, err
	}
	return data, nil
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

type PatchData struct {
	url     string
	payload interface{}
	timeout time.Duration
}

func NewPatchData(url string, payload interface{}, timeout time.Duration) *PatchData {
	return &PatchData{url: url, payload: payload, timeout: timeout}
}

func (t PatchData) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, err := request.PatchJSON(ctx, logger, t.url, t.payload, t.timeout)
	return data, err
}

// type GetPageWithOwnProxy struct {
// 	url string
// }

// func NewGetPageWithOwnProxy(url string) *GetPageWithOwnProxy {
// 	return &GetPageWithOwnProxy{url: url}
// }

// func (t GetPageWithOwnProxy) Process(ctx context.Context, logger *zap.Logger) (any, error) {
// 	tmp, err := pkg.FuncWrapper(ctx, logger, 3, 5*time.Second, NewGetRequest(t.url))
// 	if err != nil {
// 		return nil, err
// 	}
// 	userAgentResponse, ok := tmp.(*uagent.UserAgentResponse)
// 	if !ok {
// 		return nil, errors.New("parse error *model.UserAgentResponse")
// 	}
// 	req := &request.Requester{}
// 	data, err := uagent.GetPage(ctx, logger, t.url, userAgentResponse, req)
// 	return data, err
// }
