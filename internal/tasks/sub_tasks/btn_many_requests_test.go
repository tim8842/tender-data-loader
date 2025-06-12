package subtasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/util/wrappers"
)

func TestBtnaManyRequests(t *testing.T) {
	type testCase struct {
		name         string
		ids          []string
		setupMock    func()
		expectedErr  bool
		expectedData []*model.AgreementParesedData
	}

	// Сохраняем оригинал, чтобы восстановить после тестов
	originalFuncWrapper := funcWrapper

	tests := []testCase{
		{
			name: "Success with one ID",
			ids:  []string{"id1"},
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn wrappers.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 1, 4, 7:
						return &model.UserAgentResponse{}, nil
					case 2:
						return []byte("web-page"), nil
					case 3:
						return &model.AgreementParesedData{
							Pfid: "pfid1",
							Customer: &model.Customer{
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
			expectedData: []*model.AgreementParesedData{{
				ID:   "id1",
				Pfid: "pfid1",
				Customer: &model.Customer{
					ID: "cust1",
				},
			}},
		},
		{
			name: "Fail to get proxy",
			ids:  []string{"id1"},
			setupMock: func() {
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn wrappers.FuncInWrapp) (any, error) {
					return nil, errors.New("proxy error")
				}
			},
			expectedErr:  true,
			expectedData: nil,
		},
		{
			name: "Invalid type returned during parse",
			ids:  []string{"id1"},
			setupMock: func() {
				call := 0
				funcWrapper = func(ctx context.Context, logger *zap.Logger, maxRetries int, delay time.Duration, fn wrappers.FuncInWrapp) (any, error) {
					call++
					switch call {
					case 1:
						return &model.UserAgentResponse{}, nil
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

			result, err := BtnaManyRequests(ctx, logger, cfg, tt.ids)

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
