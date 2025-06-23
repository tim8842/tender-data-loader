package task

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/contract"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/supplier"
	contractt "github.com/tim8842/tender-data-loader/internal/task/contract"
	uagentt "github.com/tim8842/tender-data-loader/internal/task/uagent"
	variablet "github.com/tim8842/tender-data-loader/internal/task/variable"
	"github.com/tim8842/tender-data-loader/internal/uagent"
	"github.com/tim8842/tender-data-loader/internal/variable"
	"github.com/tim8842/tender-data-loader/pkg/parser"
	"go.uber.org/zap"
)

type BackToNowContractTask struct {
	cfg         *config.Config
	contrRepo   contract.IContractRepo
	varRepo     variable.IVariableRepo
	custRepo    customer.ICustomerRepo
	suppRepo    supplier.ISupplierRepo
	varId       string
	staticProxy bool
}

func NewBackToNowContractTask(
	cfg *config.Config,
	contrRepo contract.IContractRepo, varRepo variable.IVariableRepo,
	custRepo customer.ICustomerRepo,
	suppRepo supplier.ISupplierRepo,
	varId string,
	staticProxy bool,
) *BackToNowContractTask {
	return &BackToNowContractTask{
		cfg: cfg, contrRepo: contrRepo, varRepo: varRepo,
		custRepo: custRepo, suppRepo: suppRepo,
		varId:       varId,
		staticProxy: staticProxy,
	}
}

type StatusPayload struct {
	Status int `json:"status"`
}

func (t *BackToNowContractTask) Process(ctx context.Context, logger *zap.Logger) error {
	var mainErr error = nil
outer:
	for {
		time.Sleep(5 * time.Second)
		select {
		case <-ctx.Done():
			logger.Info("BackToNowContractTask: Context cancelled, exiting.")
			return ctx.Err()
		default:
			// time.Sleep(20 * time.Second)
			var tmp any
			var ok bool
			var err error
			var tmpByte []byte
			// Считываем данные для того чтобы хранить стейт в бд
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, variablet.NewGetVariableBackToNowContractById(t.varRepo, t.varId))
			if err != nil {
				mainErr = err
				continue outer
			}
			varData, ok := tmp.(*variable.VariableBackToNowContract)
			if !ok {
				logger.Error("Parse error *variable.VariableBackToNowContract")
				mainErr = errors.New("parse error *variable.VariableBackToNowContract")
				break outer
			}
			dbDate := parser.FromTimeToDate(varData.Vars.SignedAt)
			if parser.DateOnly(Now()) == parser.DateOnly(varData.Vars.SignedAt) {
				break outer
			}
			userAgentResponse := &uagent.UserAgentResponse{UserAgent: map[string]any{"agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.1 Safari/605.1.15"}, Proxy: map[string]any{"url": nil}} // Если часто таймаут на серваке
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
			urlNumbersPage := fmt.Sprintf(
				t.cfg.UrlZakupkiContractGetNumbers,
				varData.Vars.Fz, varData.Vars.PriceFrom, varData.Vars.PriceTo,
				dbDate, dbDate, varData.Vars.Page,
			)
			// // Получаем страницу с номерами
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, uagentt.NewGetPage(urlNumbersPage, userAgentResponse))
			if err != nil {
				logger.Error("Get numbers page error ", zap.Error(err))
				mainErr = err
				if strings.Contains(err.Error(), "неверный статус ответа: 404") ||
					strings.Contains(err.Error(), "неверный статус ответа: 5") {
					continue outer
				}
				if !t.staticProxy {
					status := 500
					if strings.Contains(err.Error(), "неверный статус ответа: 429") {
						status = 429
					}
					_, err = funcWrapper(ctx, logger, 3, 1*time.Second, uagentt.NewPatchData(
						fmt.Sprintf(t.cfg.UrlPatchProxyUsers, userAgentResponse.ID),
						&StatusPayload{Status: status},
						5*time.Second,
					))
					if err != nil {
						mainErr = err
						break outer
					}
				}
				continue outer
			}
			tmpByte, ok = tmp.([]byte)
			if !ok {
				logger.Error("Parse error []byte")
				mainErr = errors.New("parse error []byte")
				break outer
			}
			// Парсим страницу с номерами
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, contractt.NewParseData(tmpByte, contract.ParseContractIds))
			if err != nil {
				logger.Error("ParseContractIds error ", zap.Error(err))
				mainErr = err
				break outer
			}
			ids, ok := tmp.([]string)
			if !ok {
				logger.Error("Parse error []string")
				mainErr = errors.New("parse error []string")
				break outer
			}
			nextDate := func() uint8 {
				varData.Vars.SignedAt = varData.Vars.SignedAt.Add(24 * time.Hour)
				varData.Vars.Page = 1
				vData, err := varData.ConvertToVariable()
				if err != nil {
					logger.Error("Error ConvertToVariable ", zap.Error(err))
					mainErr = err
					return 1
				}
				err = t.varRepo.Update(ctx, varData.ID, &vData)
				if err != nil {
					logger.Error("Update data SignedAt сontract error ", zap.Error(err))
					mainErr = err
					return 2
				}
				return 2
			}
			if len(ids) == 0 {
				if nextDate() == 1 {
					break outer
				} else {
					continue outer
				}
			}
			// Запускаем подзадачу, которая делает параллельные 50 запросов и парсит данные
			tmp, err = funcWrapper(ctx, logger, 0, 0*time.Second, NewSBtncManyReuests(t.cfg, ids, varData.Vars.Fz, nil))
			if err != nil {
				logger.Error("Error subtasks.NewBtnсManyRequests", zap.Error(err))
				if strings.Contains(err.Error(), "no correct data, empty") {
					if nextDate() == 1 {
						break outer
					}
					continue outer
				}
				mainErr = err
				continue outer
			}
			arrData, ok := tmp.([]*contract.ContractParesedData)
			if !ok {
				logger.Error("Error parse []*model.ContractParesedData")
				mainErr = errors.New("error parse []*model.ContractParesedData")
				break outer
			}
			var customers []*customer.Customer
			var contracts []*contract.Contract
			var suppliers []*supplier.Supplier
			for _, v := range arrData {
				a, c, s := contract.ParseContractParsedDataToModel(v)
				customers = append(customers, c)
				contracts = append(contracts, a)
				suppliers = append(suppliers, s...)
			}
			err = t.contrRepo.BulkMergeMany(ctx, contracts)
			if err != nil {
				logger.Error("Error create many ", zap.Error(err))
				mainErr = err
				continue outer
			}
			err = t.suppRepo.BulkMergeMany(ctx, suppliers)
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
				logger.Error("Update data Page сontract error ", zap.Error(err))
				mainErr = err
				continue outer
			}
		}
	}
	return mainErr
}
