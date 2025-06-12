package subtasks

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/model"
	"go.uber.org/zap"
)

type ParseData struct {
	data      []byte
	parseFunc func(ctx context.Context, logger *zap.Logger, data []byte) (any, error)
}

func NewParseData(data []byte, parseFunc func(ctx context.Context, logger *zap.Logger, data []byte) (any, error)) *ParseData {
	return &ParseData{data: data, parseFunc: parseFunc}
}

func (t ParseData) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := t.parseFunc(ctx, logger, t.data)
	if ok != nil {
		return nil, ok
	}
	return data, ok
}

type ParseDataInAgreementParesedData struct {
	data      []byte
	parseFunc func(
		ctx context.Context, logger *zap.Logger, data []byte,
		agreementParesedData *model.AgreementParesedData) (any, error)
	agreementParesedData *model.AgreementParesedData
}

func NewParseDataInAgreementParesedData(data []byte, parseFunc func(
	ctx context.Context, logger *zap.Logger, data []byte,
	agreementParesedData *model.AgreementParesedData) (any, error),
	agreementParesedData *model.AgreementParesedData) *ParseDataInAgreementParesedData {
	return &ParseDataInAgreementParesedData{
		data: data, parseFunc: parseFunc, agreementParesedData: agreementParesedData,
	}
}

func (t ParseDataInAgreementParesedData) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := t.parseFunc(ctx, logger, t.data, t.agreementParesedData)
	if ok != nil {
		return nil, ok
	}
	return data, ok
}
