package tasks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tim8842/tender-data-loader/internal/config"
	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	backtonowagreementservice "github.com/tim8842/tender-data-loader/internal/service/back_to_now_agreement_service"
	subtasks "github.com/tim8842/tender-data-loader/internal/tasks/sub_tasks"
	baseutils "github.com/tim8842/tender-data-loader/internal/util/base_utils"
	"github.com/tim8842/tender-data-loader/internal/util/wrappers"
	"go.uber.org/zap"
)

type BackToNowAgreementTask struct {
	cfg          *config.Config
	repositories *repository.Repositories
	staticProxy  bool
}

var funcWrapper = wrappers.FuncWrapper
var Now = func() time.Time {
	return time.Now()
}

func NewBackToNowAgreementTask(
	cfg *config.Config,
	repositories *repository.Repositories,
	staticProxy bool,
) *BackToNowAgreementTask {
	return &BackToNowAgreementTask{cfg: cfg, repositories: repositories, staticProxy: staticProxy}
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
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, subtasks.NewGetVariableBackToNowAgreementById(t.repositories.VarRepo, "back_to_now_agreement"))
			if err != nil {
				mainErr = err
				continue outer
			}

			varData, ok := tmp.(*model.VariableBackToNowAgreement)
			if !ok {
				logger.Error("Parse error *model.VariableBackToNowAgreement")
				mainErr = errors.New("parse error *model.VariableBackToNowAgreement")
				break outer
			}
			dbDate := baseutils.FormatDate(varData.Vars.SignedAt)
			if baseutils.DateOnly(Now()) == baseutils.DateOnly(varData.Vars.SignedAt) {
				break outer
			}
			userAgentResponse := &model.UserAgentResponse{UserAgent: map[string]any{"agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36"}, Proxy: map[string]any{"url": nil}} // Если часто таймаут на серваке
			urlProx := t.cfg.UrlGetProxy
			// Получаем прокси
			if !t.staticProxy {
				tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, subtasks.NewGetRequest(urlProx))
				if err != nil {
					logger.Error("Get proxy error ", zap.Error(err))
					mainErr = err
					continue outer
				}
				userAgentResponse, ok = tmp.(*model.UserAgentResponse)
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
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, subtasks.NewGetPage(urlNumbersPage, userAgentResponse))
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
			tmp, err = funcWrapper(ctx, logger, 3, 5*time.Second, subtasks.NewParseData(tmpByte, backtonowagreementservice.ParseAgreementIds))
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
				err = t.repositories.VarRepo.Update(ctx, varData.ID, &vData)
				if err != nil {
					logger.Error("Update data SignedAt agreement error ", zap.Error(err))
					mainErr = err
					continue outer
				}
				continue outer
			}
			// Запускаем подзадачу, которая делает параллельные 50 запросов и парсит данные
			tmp, err = funcWrapper(ctx, logger, 0, 0*time.Second, subtasks.NewBtnaManyRequests(t.cfg, ids, true))
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
					err = t.repositories.VarRepo.Update(ctx, varData.ID, &vData)
					if err != nil {
						logger.Error("Update data SignedAt agreement error ", zap.Error(err))
						mainErr = err
						continue outer
					}
				}
				continue outer
			}
			arrData, ok := tmp.([]*model.AgreementParesedData)
			if !ok {
				logger.Error("Error parse []*model.AgreementParesedData")
				mainErr = errors.New("error parse []*model.AgreementParesedData")
				break outer
			}
			var customers []*model.Customer
			var agreements []*model.Agreement
			for _, v := range arrData {
				a, c := model.ParseAgreementDataToModels(v)
				customers = append(customers, c)
				agreements = append(agreements, a)
			}
			err = t.repositories.AgreementRepo.BulkMergeMany(ctx, agreements)
			if err != nil {
				logger.Error("Error create many ", zap.Error(err))
				mainErr = err
				continue outer
			}

			err = t.repositories.CustomerRepo.BulkMergeMany(ctx, customers)
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
			err = t.repositories.VarRepo.Update(ctx, varData.ID, &vData)
			if err != nil {
				logger.Error("Update data Page agreement error ", zap.Error(err))
				mainErr = err
				continue outer
			}
		}
	}
	return mainErr
}
