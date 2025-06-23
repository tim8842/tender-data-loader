package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/contract"
	contractt "github.com/tim8842/tender-data-loader/internal/task/contract"
	uagentt "github.com/tim8842/tender-data-loader/internal/task/uagent"
	"github.com/tim8842/tender-data-loader/internal/uagent"
	"go.uber.org/zap"
)

type SBtncManyRequests struct {
	cfg               *config.Config
	ids               []string
	fz                string
	userAgentResponse *uagent.UserAgentResponse
}

func NewSBtncManyReuests(cfg *config.Config, ids []string, fz string, userAgentResponse *uagent.UserAgentResponse) *SBtncManyRequests {
	return &SBtncManyRequests{cfg, ids, fz, userAgentResponse}
}

func (t SBtncManyRequests) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := BtncManyRequests(ctx, logger, t.cfg, t.ids, t.fz, t.userAgentResponse)
	if ok != nil {
		return nil, ok
	}
	return data, ok
}

func getProxy(ctx context.Context, logger *zap.Logger, url string) (*uagent.UserAgentResponse, error) {
	tmp, err := funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetRequest(url))
	if err != nil {
		return nil, err
	}
	userAgentResponse, ok := tmp.(*uagent.UserAgentResponse)
	if !ok {
		return nil, errors.New("parse error *model.UserAgentResponse")
	}
	return userAgentResponse, nil
}

func BtncManyRequests(ctx context.Context, logger *zap.Logger, cfg *config.Config, ids []string, fz string, userAgentResponse *uagent.UserAgentResponse) (any, error) {
	lenNums := len(ids)
	var wg sync.WaitGroup
	urlProx := cfg.UrlGetProxy
	ctx, cancel := context.WithCancel(ctx)
	results := make(chan *contract.ContractParesedData, lenNums)
	var res []*contract.ContractParesedData
	var mainErr error
	proxyWithGet := func(err error) bool {
		if userAgentResponse != nil || strings.Contains(err.Error(), "неверный статус ответа: 404") ||
			strings.Contains(err.Error(), "неверный статус ответа: 5") {
			if strings.Contains(err.Error(), "неверный статус ответа: 429") ||
				strings.Contains(err.Error(), "context deadline") ||
				strings.Contains(err.Error(), "неверный статус ответа: 5") { // Можно удалить чтобы 500 ошибки не ломали прогу
				mainErr = err
			}
			return true
		} else {
			status := 500
			if strings.Contains(err.Error(), "неверный статус ответа: 429") {
				status = 429
			}
			_, err = funcWrapper(ctx, logger, 3, 1*time.Second, uagentt.NewPatchData(
				fmt.Sprintf(`http://127.0.0.1:8000/api/v1/users/%d/status/`, userAgentResponse.ID),
				&StatusPayload{Status: status},
				5*time.Second,
			))
			if err != nil {
				mainErr = err
				return true
			}
		}
		return false
	}

	defer cancel()
	for i := 0; i < lenNums; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				logger.Info(fmt.Sprintf("Request № %d: Context cancelled, exiting.", i))
				return
			default:
				var err error
				var tmp any
				//получаем прокси
				// userAgentResponse := &uagent.UserAgentResponse{UserAgent: map[string]any{"agent": "Mozilla/5.0 (Linux; Android 14; Pixel 7 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.111 Mobile Safari/537.36"}, Proxy: map[string]any{"url": nil}}
				for {
					userAgentResponseWeb := userAgentResponse
					if userAgentResponse == nil {
						if userAgentResponseWeb, err = getProxy(ctx, logger, urlProx); err != nil {
							mainErr = err
							logger.Error(err.Error())
							continue
						}
					}
					// получаем web страницы
					urlWebPage := cfg.UrlZakupkiContractGetWeb + ids[i]
					tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlWebPage, userAgentResponseWeb))
					if err != nil {
						logger.Error("Contract web get err ", zap.Error(err))
						if proxyWithGet(err) {
							return
						}
					} else {
						break
					}
				}
				tmpByte, ok := tmp.([]byte)
				if !ok {
					mainErr = errors.New("parse error []byte")
					return
				}
				// Парсим эти страницы
				tmp, err = funcWrapper(ctx, logger, 0, 0, contractt.NewParseData(tmpByte, contract.ParseContractFromMain))
				if err != nil {
					if !strings.Contains(err.Error(), "error contract have no noticeId") {
						mainErr = err
					}
					return
				}
				data, ok := tmp.(*contract.ContractParesedData)
				if !ok {
					mainErr = errors.New("parse error model.ContractParesedData")
					return
				}
				data.Law = fz
				// получаем прокси
				// userAgentResponse = &uagent.UserAgentResponse{UserAgent: map[string]any{"agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"}, Proxy: map[string]any{"url": nil}}
				for {
					userAgentResponseHtml := userAgentResponse
					if userAgentResponse == nil {
						if userAgentResponseHtml, err = getProxy(ctx, logger, urlProx); err != nil {
							mainErr = err
							return
						}
					}
					// получаем html show
					urlShowHtml := cfg.UrlZakupkiContractGetHtml + data.ID
					tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlShowHtml, userAgentResponseHtml))
					if err != nil {
						if proxyWithGet(err) {
							return
						}
					} else {
						break
					}
				}
				tmpByte, ok = tmp.([]byte)
				if !ok {
					mainErr = errors.New("parse error []byte")
					return
				}
				// Парсим show
				_, err = funcWrapper(ctx, logger, 1, 5*time.Second, contractt.NewParseDataInContractParesedData(tmpByte, contract.ParseContractFromHtml, data))
				if err != nil {
					mainErr = err
					return
				}
				// получаем прокси
				// userAgentResponse = &uagent.UserAgentResponse{UserAgent: map[string]any{"agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"}, Proxy: map[string]any{"url": nil}}
				for {
					userAgentResponseCust := userAgentResponse
					if userAgentResponse == nil {
						if userAgentResponseCust, err = getProxy(ctx, logger, urlProx); err != nil {
							mainErr = err
							return
						}
					}
					urlCustomerWeb := cfg.UrlZakupkiContractGetCustomerWeb + data.Customer.ID
					tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlCustomerWeb, userAgentResponseCust))
					if err != nil {
						logger.Error("Customer get err ", zap.Error(err))
						if proxyWithGet(err) {
							return
						}
					} else {
						break
					}
				}
				tmpByte, ok = tmp.([]byte)
				if !ok {
					mainErr = errors.New("parse error []byte")
					return
				}
				// Парсим Customer
				_, err = funcWrapper(ctx, logger, 1, 5*time.Second, contractt.NewParseDataInContractParesedData(tmpByte, contract.ParseCustomerFromMain, data))
				if err != nil {
					mainErr = err
					return
				}
				// получаем прокси
				for {
					userAgentResponseCustWebAddInfo := userAgentResponse
					if userAgentResponse == nil {
						if userAgentResponseCustWebAddInfo, err = getProxy(ctx, logger, urlProx); err != nil {
							mainErr = err
							return
						}
					}
					// Получаем страницу доп инфы про customer
					urlCustomerWebAddInfo := fmt.Sprintf(cfg.UrlZakupkiContractGetCustomerWebAddinfo, data.Customer.ID)
					tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlCustomerWebAddInfo, userAgentResponseCustWebAddInfo))
					if err != nil {
						logger.Error("Customer get err ", zap.Error(err))
						if proxyWithGet(err) {
							return
						}
					} else {
						break
					}
				}
				tmpByte, ok = tmp.([]byte)
				if !ok {
					mainErr = errors.New("parse error []byte")
					return
				}
				// Парсим Customer
				_, err = funcWrapper(ctx, logger, 1, 5*time.Second, contractt.NewParseDataInContractParesedData(tmpByte, contract.ParseCustomerFromMainAddInfo, data))
				if err != nil {
					mainErr = err
					return
				}
				results <- data
			}

		}()
	}
	wg.Wait()
	close(results)
	for msg := range results {
		res = append(res, msg)
	}

	if mainErr != nil {
		return nil, mainErr
	}
	if len(res) == 0 {
		logger.Error("No correct data, empty")
		mainErr = errors.New("no correct data, empty")
	}
	logger.Info("Main: all requests finished")
	return res, mainErr
}
