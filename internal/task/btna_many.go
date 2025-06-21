package task

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/config"

	agreementt "github.com/tim8842/tender-data-loader/internal/task/agreement"
	uagentt "github.com/tim8842/tender-data-loader/internal/task/uagent"
	"github.com/tim8842/tender-data-loader/internal/uagent"
	"go.uber.org/zap"
)

type SBtnaManyRequests struct {
	cfg         *config.Config
	ids         []string
	staticProxy bool
}

func NewBtnaManyRequests(cfg *config.Config, ids []string, staticProxy bool) *SBtnaManyRequests {
	return &SBtnaManyRequests{cfg: cfg, ids: ids, staticProxy: staticProxy}
}

func (t SBtnaManyRequests) Process(ctx context.Context, logger *zap.Logger) (any, error) {
	data, ok := BtnaManyRequests(ctx, logger, t.cfg, t.ids, t.staticProxy)
	if ok != nil {
		return nil, ok
	}
	return data, ok
}

func BtnaManyRequests(ctx context.Context, logger *zap.Logger, cfg *config.Config, ids []string, staticProxy bool) (any, error) {
	lenNums := len(ids)
	results := make(chan *agreement.AgreementParesedData, lenNums)
	var res []*agreement.AgreementParesedData
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	var mainErr error = nil
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
				var tmp any
				var ok bool
				var err error
				var tmpByte []byte
				userAgentResponse := &uagent.UserAgentResponse{UserAgent: map[string]any{"agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"}, Proxy: map[string]any{"url": nil}} // Если часто таймаут на серваке
				urlProx := cfg.UrlGetProxy
				if !staticProxy {
					// Получаем прокси
					tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetRequest(urlProx))
					if err != nil {
						mainErr = err
						return
					}
					userAgentResponse, ok = tmp.(*uagent.UserAgentResponse)
					if !ok {
						mainErr = errors.New("parse error *model.UserAgentResponse")
						return
					}
				}

				// Делаем запрос по каждому id
				urlWebPage := cfg.UrlZakupkiAgreementGetAgreegmentWeb + ids[i]
				tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlWebPage, userAgentResponse))
				if err != nil {
					logger.Error("Agreement web get err ", zap.Error(err))
					mainErr = err
					return
				}
				tmpByte, ok = tmp.([]byte)
				if !ok {
					mainErr = errors.New("parse error []byte")
					return
				}
				// Парсим эти страницы
				tmp, _ = funcWrapper(ctx, logger, 1, 5*time.Second, agreementt.NewParseData(tmpByte, agreement.ParseAgreementFromMain))
				if tmp == nil {
					logger.Error("Parse error no noticed id")
					mainErr = err
					return
				}
				data, ok := tmp.(*agreement.AgreementParesedData)
				if !ok {
					mainErr = errors.New("parse error model.AgreementParesedData")
					return
				}
				data.ID = ids[i]
				// Получаем прокси
				if !staticProxy {
					tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetRequest(urlProx))
					if err != nil {
						mainErr = err
						return
					}
					userAgentResponse, ok = tmp.(*uagent.UserAgentResponse)
					if !ok {
						mainErr = errors.New("parse error *model.UserAgentResponse")
						return
					}
				}
				// Получаем html show
				if data.Pfid != "" {
					urlShowHtml := cfg.UrlZakupkiAgreementGetAgreegmentShowHtml + data.Pfid
					tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlShowHtml, userAgentResponse))
					if err != nil {
						logger.Error("Agreement html show get err ", zap.Error(err))
						mainErr = err
						return
					}
					tmpByte, ok = tmp.([]byte)
					if !ok {
						mainErr = errors.New("parse error []byte")
						return
					}
					// Парсим show
					_, err = funcWrapper(ctx, logger, 1, 5*time.Second, agreementt.NewParseDataInAgreementParesedData(tmpByte, agreement.ParseAgreementFromHtml, data))
					if err != nil {
						mainErr = err
						return
					}
				}
				// _, ok = tmp.(*model.AgreementParesedData)
				// if !ok {
				// 	logger.Error("Parse error model.AgreementParesedData", zap.Error(err))
				// 	mainErr = err
				// 	return
				// }
				// Получаем прокси
				if !staticProxy {
					tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetRequest(urlProx))
					if err != nil {
						mainErr = err
						return
					}
					userAgentResponse, ok = tmp.(*uagent.UserAgentResponse)
					if !ok {
						mainErr = errors.New("parse error *model.UserAgentResponse")
						return
					}
				}
				// Получаем страницу customer
				urlCustomerWeb := cfg.UrlZakupkiAgreementGetCustomerWeb + data.Customer.ID
				tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlCustomerWeb, userAgentResponse))
				if err != nil {
					logger.Error("Customer get err ", zap.Error(err))
					mainErr = err
					return
				}
				tmpByte, ok = tmp.([]byte)
				if !ok {
					mainErr = errors.New("parse error []byte")
					return
				}
				// Парсим Customer
				_, err = funcWrapper(ctx, logger, 1, 5*time.Second, agreementt.NewParseDataInAgreementParesedData(tmpByte, agreement.ParseCustomerFromMain, data))
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
