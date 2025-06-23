package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/contract"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/uagent"
	"github.com/tim8842/tender-data-loader/pkg"

	"go.uber.org/zap"
)

func TestBtncManyRequests(t *testing.T) {
	type testCase struct {
		name         string
		ids          []string
		staticProxy  *uagent.UserAgentResponse
		setupMock    func()
		expectedErr  bool
		expectedData []*contract.ContractParesedData
	}

	// Сохраняем оригинал, чтобы восстановить после тестов
	originalFuncWrapper := funcWrapper

	tests := []testCase{
		{
			name:        "Success with one ID",
			ids:         []string{"id1"},
			staticProxy: nil,
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 1, 4, 7, 10:
						return &uagent.UserAgentResponse{}, nil
					case 2:
						return []byte("web-page"), nil
					case 3:
						return &contract.ContractParesedData{
							ID:       "1234",
							Customer: &customer.Customer{ID: "ID"},
						}, nil
					case 5:
						return []byte("show-page"), nil
					case 6, 9, 12:
						return nil, nil
					case 8, 11:
						return []byte("customer-page"), nil
					default:
						return nil, errors.New("unexpected call")
					}
				}
			},
			expectedErr: false,
			expectedData: []*contract.ContractParesedData{{
				ID:       "1234",
				Law:      "fz44",
				Customer: &customer.Customer{ID: "ID"},
			}},
		},
		{
			name:        "Success with one ID static",
			ids:         []string{"id1"},
			staticProxy: &uagent.UserAgentResponse{},
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 2:
						return &contract.ContractParesedData{
							ID:       "1234",
							Customer: &customer.Customer{ID: "ID"},
						}, nil
					case 1, 3, 5, 7:
						return []byte("page"), nil
					case 4, 6, 8:
						return nil, nil
					default:
						return nil, errors.New("unexpected call")
					}
				}
			},
			expectedErr: false,
			expectedData: []*contract.ContractParesedData{{
				ID:       "1234",
				Law:      "fz44",
				Customer: &customer.Customer{ID: "ID"},
			}},
		},
		{
			name:        "fail get proxy",
			ids:         []string{"id1"},
			staticProxy: nil,
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 1:
						return nil, errors.New("")
					default:
						return nil, errors.New("unexpected call")
					}
				}
			},
			expectedErr:  true,
			expectedData: []*contract.ContractParesedData{},
		},
		{
			name:        "fail get parse type from parsed",
			ids:         []string{"id1"},
			staticProxy: nil,
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 1:
						return &uagent.UserAgentResponse{}, nil
					case 2:
						return []byte("page"), nil
					case 3:
						return map[string]any{"hello": 1}, nil
					default:
						return nil, errors.New("unexpected call")
					}
				}
			},
			expectedErr:  true,
			expectedData: []*contract.ContractParesedData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				funcWrapper = originalFuncWrapper
			}()

			tt.setupMock()

			logger := zap.NewNop()
			ctx := context.Background()
			cfg := &config.Config{
				UrlGetProxy:                             "proxy/",
				UrlZakupkiContractGetWeb:                "web/",
				UrlZakupkiContractGetHtml:               "show/",
				UrlZakupkiContractGetCustomerWeb:        "cust/",
				UrlZakupkiContractGetCustomerWebAddinfo: "test/",
			}

			result, err := BtncManyRequests(ctx, logger, cfg, tt.ids, "fz44", tt.staticProxy)

			if tt.expectedErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}
		})
	}
}
