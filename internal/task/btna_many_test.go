package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/uagent"
	"github.com/tim8842/tender-data-loader/pkg"

	"go.uber.org/zap"
)

func TestBtnaManyRequests(t *testing.T) {
	type testCase struct {
		name         string
		ids          []string
		staticProxy  bool
		setupMock    func()
		expectedErr  bool
		expectedData []*agreement.AgreementParesedData
	}

	// Сохраняем оригинал, чтобы восстановитьd после тестов
	originalFuncWrapper := funcWrapper

	tests := []testCase{
		{
			name:        "Success with one ID",
			ids:         []string{"id1"},
			staticProxy: false,
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 1, 4, 7:
						return &uagent.UserAgentResponse{}, nil
					case 2:
						return []byte("web-page"), nil
					case 3:
						return &agreement.AgreementParesedData{
							Pfid: "pfid1",
							Customer: &customer.Customer{
								ID: "cust1",
							},
						}, nil
					case 5:
						return []byte("show-page"), nil
					case 6:
						return nil, nil
					case 8:
						return []byte("customer-page"), nil
					case 9:
						return nil, nil
					default:
						return nil, errors.New("unexpected call")
					}
				}
			},
			expectedErr: false,
			expectedData: []*agreement.AgreementParesedData{{
				ID:   "id1",
				Pfid: "pfid1",
				Customer: &customer.Customer{
					ID: "cust1",
				},
			}},
		},
		{
			name:        "Success with one ID static",
			ids:         []string{"id1"},
			staticProxy: true,
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 1:
						return []byte("web-page"), nil
					case 2:
						return &agreement.AgreementParesedData{
							Pfid: "pfid1",
							Customer: &customer.Customer{
								ID: "cust1",
							},
						}, nil
					case 3:
						return []byte("show-page"), nil
					case 4:
						return nil, nil
					case 5:
						return []byte("customer-page"), nil
					case 6:
						return nil, nil
					default:
						return nil, errors.New("unexpected call")
					}
				}
			},
			expectedErr: false,
			expectedData: []*agreement.AgreementParesedData{{
				ID:   "id1",
				Pfid: "pfid1",
				Customer: &customer.Customer{
					ID: "cust1",
				},
			}},
		},
		{
			name:        "Fail to get proxy",
			staticProxy: false,
			ids:         []string{"id1"},
			setupMock: func() {
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
					return nil, errors.New("proxy error")
				}
			},
			expectedErr:  true,
			expectedData: nil,
		},
		{
			name:        "Invalid type returned during parse",
			ids:         []string{"id1"},
			staticProxy: false,
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn pkg.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 1:
						return &uagent.UserAgentResponse{}, nil
					case 2:
						return []byte("html"), nil
					case 3:
						return "not *AgreementParesedData", nil
					default:
						return nil, errors.New("should not be called")
					}
				}
			},
			expectedErr:  true,
			expectedData: nil,
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
				UrlGetProxy:                              "proxy/",
				UrlZakupkiAgreementGetAgreegmentWeb:      "web/",
				UrlZakupkiAgreementGetAgreegmentShowHtml: "show/",
				UrlZakupkiAgreementGetCustomerWeb:        "cust/",
			}

			result, err := BtnaManyRequests(ctx, logger, cfg, tt.ids, tt.staticProxy)

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
