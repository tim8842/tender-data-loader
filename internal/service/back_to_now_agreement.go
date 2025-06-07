package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/tim8842/tender-data-loader/internal/model"

	"github.com/tim8842/tender-data-loader/internal/util/base"
	"github.com/tim8842/tender-data-loader/internal/util/requests"
	"go.uber.org/zap"
)

type UserAgentResponse struct {
	ID        int            `json:"id"`
	Proxy     map[string]any `json:"proxy"`
	Status    int            `json:"status"`
	UpdatedAt time.Time      `json:"updated_at"`
	UserAgent map[string]any `json:"user_agent"`
}

type PageInner struct {
	UserAgent UserAgentResponse `json:"user_agent"`
	Url       string            `json:"url"`
}

func GetUserAgent(ctx context.Context, url string, logger *zap.Logger) (*UserAgentResponse, error) {
	var usStruct UserAgentResponse
	res, err := requests.GetRequest(ctx, url, 5*time.Second, logger)
	if err != nil {
		return nil, err
	}
	ok := json.Unmarshal(res, &usStruct)
	if ok != nil {
		logger.Info("Не может привести тип ответа")
		return nil, ok
	}
	return &usStruct, err
}

func GetPage(ctx context.Context, inner PageInner, logger *zap.Logger) (interface{}, error) {
	userAgent, proxyUrl := "", ""
	if tmp, ok := inner.UserAgent.UserAgent["agent"].(string); ok {
		userAgent = tmp
	}
	if tmp, ok := inner.UserAgent.Proxy["url"].(string); ok {
		proxyUrl = tmp
	}
	res, err := requests.GetRequest(ctx, inner.Url, 5*time.Second, logger, requests.RequestOptions{UserAgent: userAgent, ProxyUrl: proxyUrl})
	if err != nil {
		return nil, err
	}
	return res, err
}

func ParseAgreementIds(ctx context.Context, inner []byte, logger *zap.Logger) ([]string, error) {
	reader := bytes.NewReader(inner)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		logger.Fatal("err parsing numbers" + err.Error())
		return nil, err
	}
	var ids []string
	doc.Find(".registry-entry__header-mid__number a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		// href может быть относительным, можно дополнительно обработать, если надо

		// парсим href, чтобы достать параметр id
		id := getParamFromHref(href, "id")

		if id != "" {
			ids = append(ids, id)
		}
	})
	if len(ids) > 0 {
		return ids, nil
	} else {
		return nil, errors.New("zero ids")
	}

}

func ParseAgreementFromMain(body []byte, logger *zap.Logger) (*model.Agreement, error) {
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	data := &model.Agreement{}

	// 1. Номер (№ 82902040527250000050000 → 82902040527250000050000)
	doc.Find("span.cardMainInfo__purchaseLink.distancedText a").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		text = strings.TrimPrefix(text, "№")
		data.Number = strings.ReplaceAll(text, " ", "")
	})
	// 2. Статус
	doc.Find("span.cardMainInfo__state").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		data.Status = text
	})
	// pfid
	doc.Find(`a[title="Печатная форма"]`).Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			data.Pfid = getParamFromHref(href, "pfid")
		}
	})

	// 2. Цена
	doc.Find(".rightBlock__price").EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		text = strings.ReplaceAll(text, "₽", "")
		text = strings.ReplaceAll(text, ",", ".")
		if price, err := base.ParsePriceToFloat(text); err == nil {
			data.Price = price
			return false // найдена первая цена
		}
		return true
	})

	// 3. 4 даты по rightBlock__text
	doc.Find(".rightBlock__text").Each(func(i int, s *goquery.Selection) {
		if i == 4 {
			return
		}
		text := strings.TrimSpace(s.Text())
		text = strings.ReplaceAll(text, " ", "")
		text = strings.ReplaceAll(text, "\u00A0", "")
		texts := strings.Split(text, "—")
		for idx, v := range texts {
			text := strings.TrimSpace(v)
			date, err := base.ParseDate(text)
			if err == nil {
				switch i {
				case 0:
					data.SignedAt = date
				case 1:
					if idx == 0 {
						data.ExecutionStart = date
					} else {
						data.ExecutionEnd = date
					}
				case 2:
					data.PublishedAt = date
				case 3:
					data.UpdatedAt = date
				}
			}
		}
	})

	// 5. noticeInfoId из 3-го .section__info a
	// 6. agencyId из 9-го .section__info a
	doc.Find(".section__info a").Each(func(i int, s *goquery.Selection) {
		if i == 2 {
			if href, ok := s.Attr("href"); ok {
				data.NoticeId = getParamFromHref(href, "noticeInfoId")
			}
		}
		if i == 8 {
			if href, ok := s.Attr("href"); ok {
				data.Customer.ID = getParamFromHref(href, "agencyId")
			}
		}
	})
	doc.Find("span.cardMainInfo__content a").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			data.Customer.ID = getParamFromHref(href, "agencyId")
		}
	})
	return data, nil
}

func getParamFromHref(href, param string) string {
	parts := strings.Split(href, "?")
	if len(parts) < 2 {
		return ""
	}
	for _, p := range strings.Split(parts[1], "&") {
		if strings.HasPrefix(p, param+"=") {
			return strings.TrimPrefix(p, param+"=")
		}
	}
	return ""
}

func getFromHtmlByTitle(doc *goquery.Document, tag string, title string) (string, error) {
	result := ""
	doc.Find(tag).Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == title {
			// Ищем следующий <td>
			val := s.Parent().Find(tag).Eq(s.Index() + 1)
			result = val.Text()
		}
	})
	if result != "" {
		return result, nil
	} else {
		return "", errors.New("No element with titile" + title)
	}
}

func ParseAgreementFromHtml(body []byte, data *model.Agreement, logger *zap.Logger) (*model.Agreement, error) {
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	data.Customer.Code, _ = getFromHtmlByTitle(doc, "td", "Идентификационный код заказчика:")
	data.Customer.Name, _ = getFromHtmlByTitle(doc, "td", "Полное наименование организации:")
	data.Customer.INN, _ = getFromHtmlByTitle(doc, "td", "ИНН/КПП:")
	data.Customer.INN = strings.Split(strings.ReplaceAll(data.Customer.INN, " ", ""), "/")[0]
	data.PurchaseMethod, _ = getFromHtmlByTitle(doc, "td", "Способ закупки:")
	data.Subject, _ = getFromHtmlByTitle(doc, "td", "Предмет договора:")
	var columns []string
	doc.Find(".item-information").Eq(1).Find("tr").Each(func(i int, row *goquery.Selection) {
		// var columns []string
		row.Find("th").Each(func(j int, cell *goquery.Selection) {
			columns = append(columns, cell.Text())
		})
		if i > 0 {
			service := model.AgreementService{}
			row.Find("td").Each(func(j int, cell *goquery.Selection) {
				switch columns[j] {
				case "Наименование товаров, работ, услуг":
					service.Name = cell.Text()
				case "Классификация по ОКПД", "ОКПД(ОКДП)":
					service.OKPD = cell.Text()
				case "Классификация по ОКПД2":
					service.OKPD2 = cell.Text()
				case "Количество (Объем)":
					service.Quantity, _ = base.ParsePriceToFloat(cell.Text())
				case "Единица измерения":
					service.QuantityType = cell.Text()
				case "Количество (объем), единица измерения":
					tmp := strings.Split(cell.Text(), ",")
					service.Quantity, _ = base.ParsePriceToFloat(strings.TrimSpace(tmp[0]))
					if len(tmp) > 1 {
						service.QuantityType = strings.TrimSpace(tmp[1])
					}
				case "Цена за единицу":
					tmp := strings.Split(cell.Text(), ",")
					service.UnitPrice, _ = base.ParsePriceToFloat(strings.TrimSpace(tmp[0]))
					if len(tmp) > 1 {
						service.Currency = strings.TrimSpace(tmp[1])
					}
				case "Страна происхождения товара", "Страна происхождения (производителя) товара":
					service.CountryOfOrigin = strings.TrimSpace(cell.Text())
				case "Страна регистрации производителя товара":
					service.CountryRegistered = strings.TrimSpace(cell.Text())
				}
			})
			data.Services = append(data.Services, service)
		}
	})

	return data, nil
}

func ParseCustomerFromMain(body []byte, data *model.Agreement, logger *zap.Logger) (*model.Agreement, error) {
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	doc.Find(".registry-entry__body-value").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			data.Customer.Location = strings.TrimSpace(s.Text())
		}
	})
	data.Customer.MainWork, _ = getFromHtmlByTitle(doc, "span", "Коды основного вида деятельности по ОКВЭД")
	data.Customer.MainWork = strings.TrimSpace(data.Customer.MainWork)
	data.Customer.OKOPF, _ = getFromHtmlByTitle(doc, "span", "Код по ОКОПФ")
	data.Customer.OKOPF = strings.TrimSpace(data.Customer.OKOPF)
	logger.Debug(fmt.Sprintf("%+v\n", data))
	return data, nil
}
