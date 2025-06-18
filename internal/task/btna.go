package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/customer"
	agreementt "github.com/tim8842/tender-data-loader/internal/task/agreement"
	uagentt "github.com/tim8842/tender-data-loader/internal/task/uagent"
	variablet "github.com/tim8842/tender-data-loader/internal/task/variable"
	"github.com/tim8842/tender-data-loader/internal/uagent"
	"github.com/tim8842/tender-data-loader/internal/variable"
	"github.com/tim8842/tender-data-loader/pkg"
	"github.com/tim8842/tender-data-loader/pkg/parser"
	"go.uber.org/zap"
)

type BackToNowAgreementTask struct {
	cfg         *config.Config
	agreeRepo   agreement.IAgreementRepo
	varRepo     variable.IVariableRepo
	custRepo    customer.ICustomerRepo
	staticProxy bool
}

var funcWrapper = pkg.FuncWrapper
var Now = func() time.Time {
	return time.Now()
}

func NewBackToNowAgreementTask(
	cfg *config.Config,
	agreeRepo agreement.IAgreementRepo, varRepo variable.IVariableRepo,
	custRepo customer.ICustomerRepo,
	staticProxy bool,
) *BackToNowAgreementTask {
	return &BackToNowAgreementTask{
		cfg: cfg, agreeRepo: agreeRepo, varRepo: varRepo,
		custRepo: custRepo, staticProxy: staticProxy,
	}
}

func (t *BackToNowAgreementTask) Process(ctx context.Context, logger *zap.Logger) error {
	var mainErr error = nil
outer:
	for {
		select {
		case <-ctx.Done():
			logger.Info("BackToNowAgreementTask: Context cancelled, exiting.")
			return ctx.Err()
		default:
			var tmp any
			var ok bool
			var err error
			var tmpByte []byte
			// Считываем данные для того чтобы хранить стейт в бд
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, variablet.NewGetVariableBackToNowById(t.varRepo, "back_to_now_agreement"))
			if err != nil {
				mainErr = err
				continue outer
			}

			varData, ok := tmp.(*variable.VariableBackToNow)
			if !ok {
				logger.Error("Parse error *model.VariableBackToNow")
				mainErr = errors.New("parse error *model.VariableBackToNow")
				break outer
			}
			dbDate := parser.FromTimeToDate(varData.Vars.SignedAt)
			if parser.DateOnly(Now()) == parser.DateOnly(varData.Vars.SignedAt) {
				break outer
			}
			userAgentResponse := &uagent.UserAgentResponse{UserAgent: map[string]any{"agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"}, Proxy: map[string]any{"url": nil}} // Если часто таймаут на серваке
			urlProx := t.cfg.UrlGetProxy
			// Получаем прокси
			if !t.staticProxy {
				tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetRequest(urlProx))
				if err != nil {
					logger.Error("Get proxy error ", zap.Error(err))
					mainErr = err
					continue outer
				}
				userAgentResponse, ok = tmp.(*uagent.UserAgentResponse)
				if !ok {
					logger.Error("Parse error model.UserAgentResponse")
					mainErr = errors.New("parse error model.UserAgentResponse")
					break outer
				}
			}
			urlNumbersPage := t.cfg.UrlZakupkiAgreementGetNumbersFirst +
				dbDate +
				t.cfg.UrlZakupkiAgreementGetNumbersSecond +
				dbDate +
				t.cfg.UrlZakupkiAgreementGetNumbersThird +
				fmt.Sprintf("%d", varData.Vars.Page) +
				t.cfg.UrlZakupkiAgreementGetNumbersForth
			// // Получаем страницу с номерами
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlNumbersPage, userAgentResponse))
			if err != nil {
				logger.Error("Get numbers page error ", zap.Error(err))
				mainErr = err
				continue outer
			}
			tmpByte, ok = tmp.([]byte)
			if !ok {
				logger.Error("Parse error []byte")
				mainErr = errors.New("parse error []byte")
				break outer
			}
			// Парсим страницу с номерами
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, agreementt.NewParseData(tmpByte, agreement.ParseAgreementIds))
			if err != nil {
				logger.Error("ParseAgreementIds error ", zap.Error(err))
				mainErr = err
				break outer
			}
			ids, ok := tmp.([]string)
			if !ok {
				logger.Error("Parse error []string")
				mainErr = errors.New("parse error []string")
				break outer
			}
			if len(ids) == 0 {
				varData.Vars.SignedAt = varData.Vars.SignedAt.Add(24 * time.Hour)
				varData.Vars.Page = 1
				vData, err := varData.ConvertToVariable()
				if err != nil {
					logger.Error("Error ConvertToVariable ", zap.Error(err))
					mainErr = err
					break outer
				}
				err = t.varRepo.Update(ctx, varData.ID, &vData)
				if err != nil {
					logger.Error("Update data SignedAt agreement error ", zap.Error(err))
					mainErr = err
					continue outer
				}
				continue outer
			}
			// Запускаем подзадачу, которая делает параллельные 50 запросов и парсит данные
			tmp, err = funcWrapper(ctx, logger, 0, 0*time.Second, NewBtnaManyRequests(t.cfg, ids, true))
			if err != nil {
				logger.Error("Error subtasks.NewBtnaManyRequests", zap.Error(err))
				mainErr = err
				if strings.Contains(err.Error(), "no correct data, empty") {
					varData.Vars.SignedAt = varData.Vars.SignedAt.Add(24 * time.Hour)
					varData.Vars.Page = 1
					vData, err := varData.ConvertToVariable()
					if err != nil {
						logger.Error("Error ConvertToVariable ", zap.Error(err))
						mainErr = err
						break outer
					}
					err = t.varRepo.Update(ctx, varData.ID, &vData)
					if err != nil {
						logger.Error("Update data SignedAt agreement error ", zap.Error(err))
						mainErr = err
						continue outer
					}
				}
				continue outer
			}
			arrData, ok := tmp.([]*agreement.AgreementParesedData)
			if !ok {
				logger.Error("Error parse []*model.AgreementParesedData")
				mainErr = errors.New("error parse []*model.AgreementParesedData")
				break outer
			}
			var customers []*customer.Customer
			var agreements []*agreement.Agreement
			for _, v := range arrData {
				a, c := agreement.ParseAgreementDataToModels(v)
				customers = append(customers, c)
				agreements = append(agreements, a)
			}
			err = t.agreeRepo.BulkMergeMany(ctx, agreements)
			if err != nil {
				logger.Error("Error create many ", zap.Error(err))
				mainErr = err
				continue outer
			}

			err = t.custRepo.BulkMergeMany(ctx, customers)
			if err != nil {
				logger.Error("Error create many ", zap.Error(err))
				mainErr = err
				continue outer
			}
			if varData.Vars.Page < 100 {
				varData.Vars.Page = varData.Vars.Page + 1
			} else {
				varData.Vars.Page = 1
				varData.Vars.SignedAt = varData.Vars.SignedAt.Add(24 * time.Hour)
			}

			vData, err := varData.ConvertToVariable()
			if err != nil {
				logger.Error("Error ConvertToVariable ", zap.Error(err))
				mainErr = err
				break outer
			}
			err = t.varRepo.Update(ctx, varData.ID, &vData)
			if err != nil {
				logger.Error("Update data Page agreement error ", zap.Error(err))
				mainErr = err
				continue outer
			}
		}
	}
	return mainErr
}
