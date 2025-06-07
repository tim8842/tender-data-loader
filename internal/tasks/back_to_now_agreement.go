package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/tim8842/tender-data-loader/internal/model"
	"github.com/tim8842/tender-data-loader/internal/repository"
	"github.com/tim8842/tender-data-loader/internal/service"
	"github.com/tim8842/tender-data-loader/internal/util"
	"go.uber.org/zap"
)

func BackToNowAgreementTask(ctx context.Context, logger *zap.Logger, variableRepo repository.IMongoRepository) error {
	for {

		select {
		case <-ctx.Done():
			logger.Info("BackToNowAgreementTask: Context cancelled, exiting.")
			return ctx.Err()
		default:
			var tmp interface{}
			tmp, _ = util.FuncWrapper(ctx, func(ctx context.Context, res string, logger *zap.Logger) (interface{}, error) {
				data, ok := variableRepo.GetByID(ctx, res)
				var model model.VariableBackToNowAgreement
				b, ok := json.Marshal(data)
				ok = json.Unmarshal(b, &model)
				fmt.Println(model)
				if ok == nil {
					return model, ok
				} else {
					return nil, ok
				}
			}, 3, 5*time.Second, "back_to_now_agreement", logger)
			varData, _ := tmp.(model.VariableBackToNowAgreement)
			// Полчаем прокси
			tmp, _ = util.FuncWrapper(ctx, func(ctx context.Context, url string, logger *zap.Logger) (interface{}, error) {
				data, ok := service.GetUserAgent(ctx, url, logger)
				return data, ok
			}, 3, 5*time.Second, os.Getenv("URL_GET_PROXY"), logger)
			userAgent, _ := tmp.(service.UserAgentResponse)
			dbDate := util.FormatDate(varData.Vars.Signed_at)
			resForGetNumbers := service.PageInner{UserAgent: userAgent, Url: os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FIRST") +
				dbDate + os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_SECOND") + dbDate + os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_THIRD") + fmt.Sprintf("%d", varData.Vars.Page) + os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FORTH"),
			}

			// Получаем страницу с номерами
			tmp, _ = util.FuncWrapper(ctx, service.GetPage, 3, 5*time.Second, resForGetNumbers, logger)
			// парсим ее
			tmp, _ = util.FuncWrapper(ctx, service.ParseAgreementIds, 3, 5*time.Second, tmp.([]byte), logger)
			numbers := tmp.([]string)
			// url := res
			// получаем новый прокси для новых запросов
			tmp, _ = util.FuncWrapper(ctx, func(ctx context.Context, url string, logger *zap.Logger) (interface{}, error) {
				data, ok := service.GetUserAgent(ctx, url, logger)
				return data, ok
			}, 3, 5*time.Second, os.Getenv("URL_GET_PROXY"), logger)
			userAgent, _ = tmp.(service.UserAgentResponse)
			// resForGetNumbers := service.PageInner{UserAgent: userAgent, Url: os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FIRST") +
			// 	dbDate + os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_SECOND") + dbDate + os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_THIRD") + string(varData.Vars.Page) + os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_NUMBERS_FORTH"),
			// }
			tmp, _ = util.FuncWrapper(ctx, ManyRequests, 3, 5*time.Second, map[string]any{"numbers": numbers, "userAgent": userAgent}, logger)
			time.Sleep(5 * time.Second)
		}
	}
}

func ManyRequests(ctx context.Context, data any, logger *zap.Logger) (any, error) {
	var mainOk error
	mapa, ok := data.(map[string]any)
	if !(ok) {
		logger.Error("error parse")
		return nil, errors.New("error parse")
	}
	numbers, ok := mapa["numbers"].([]string)
	if !(ok) {
		logger.Error("error parse")
		return nil, errors.New("error parse")
	}
	userAgent, ok := mapa["userAgent"].(service.UserAgentResponse)
	if !(ok) {
		logger.Error("error parse")
		return nil, errors.New("error parse")
	}
	lenNums := len(numbers)
	results := make(chan *model.Agreement, lenNums)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for i := 0; i < lenNums; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				logger.Info(fmt.Sprintf("Request № %d: Context cancelled, exiting.", i))
			default:
				url := os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_AGREEGMENT_WEB") + numbers[i]
				page, err := service.GetPage(ctx, service.PageInner{Url: url, UserAgent: userAgent}, logger)
				if err != nil {
					mainOk = err
				}
				model, err := service.ParseAgreementFromMain(page.([]byte), logger)
				if err != nil {
					mainOk = err
				}
				model.ID = numbers[i]
				url = os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_AGREEGMENT_SHOW_HTML") + model.Pfid
				userAgent, _ := service.GetUserAgent(ctx, os.Getenv("URL_GET_PROXY"), logger)
				page, err = service.GetPage(ctx, service.PageInner{Url: url, UserAgent: *userAgent}, logger)
				if err != nil {
					return
				}
				model, err = service.ParseAgreementFromHtml(page.([]byte), model, logger)
				url = os.Getenv("URL_ZAKUPKI_AGREEMENT_GET_CUSTOMER_WEB") + model.Customer.ID
				userAgent, _ = service.GetUserAgent(ctx, os.Getenv("URL_GET_PROXY"), logger)
				page, err = service.GetPage(ctx, service.PageInner{Url: url, UserAgent: *userAgent}, logger)
				if err != nil {
					return
				}
				model, err = service.ParseCustomerFromMain(page.([]byte), model, logger)
				logger.Info(fmt.Sprintf("Request № %d: ended.", i))
				//procees func
				//results <- res
			}
		}()
	}
	wg.Wait()
	close(results)
	logger.Info("Main: all requests finished")
	return nil, mainOk
}
