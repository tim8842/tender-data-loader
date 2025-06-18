package contract

import (
	"context"

	"github.com/tim8842/tender-data-loader/internal/contract"
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

type ParseDataInContractParesedData struct {
	data      []byte
	parseFunc func(
		ctx context.Context, logger *zap.Logger, data []byte,
		contractParesedData *contract.ContractParesedData) (any, error)
	contractParesedData *contract.ContractParesedData
}

func NewParseDataInContractParesedData(data []byte, parseFunc func(
	ctx context.Context, logger *zap.Logger, data []byte,
	contractParesedData *contract.ContractParesedData) (any, error),
	contractParesedData *contract.ContractParesedData) *ParseDataInContractParesedData {
	return &ParseDataInContractParesedData{
		data: data, parseFunc: parseFunc, contractParesedData: contractParesedData,
	}
}

func (t ParseDataInContractParesedData) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := t.parseFunc(ctx, logger, t.data, t.contractParesedData)
	if ok != nil {
		return nil, ok
	}
	return data, ok
}
