package backtonowagreementservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tim8842/tender-data-loader/internal/model"
	baseutils "github.com/tim8842/tender-data-loader/internal/util/base_utils"
	"go.uber.org/zap"
)

func GetParamFromHref(href, param string) string {
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

func ParseAgreementIds(ctx context.Context, logger *zap.Logger, data []byte) (any, error) {
	reader := bytes.NewReader(data)
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
		id := GetParamFromHref(href, "id")

		if id != "" {
			ids = append(ids, id)
		}
	})
	if len(ids) > 0 {
		return ids, nil
	} else {
		return []string{}, nil
	}

}

func ParseAgreementFromMain(ctx context.Context, logger *zap.Logger, body []byte) (any, error) {
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, errors.New("failed to parse HTML: " + error.Error(err))
	}
	data := &model.AgreementParesedData{Customer: &model.Customer{}}
	noticeSpan, err := getElementFromHtmlByTitle(doc, "span", "Извещение о закупке")
	if err == nil {
		noticeSpan.Children().Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				data.NoticeId = GetParamFromHref(href, "noticeInfoId")
			}

		})
	}
	if data.NoticeId == "" {
		return nil, nil
	}
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
			data.Pfid = GetParamFromHref(href, "pfid")
		}
	})
	// 2. Цена
	doc.Find(".rightBlock__price").EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		text = strings.ReplaceAll(text, "₽", "")
		text = strings.ReplaceAll(text, ",", ".")
		if price, err := baseutils.ParsePriceToFloat(text); err == nil {
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
			date, err := baseutils.ParseDate(text)
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

	agencyIdSpan, err := getElementFromHtmlByTitle(doc, "span", "Заказчик")
	if err == nil {
		agencyIdSpan.Find("a").Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				data.Customer.ID = GetParamFromHref(href, "agencyId")
			}

		})
	}
	return data, nil
}

func ParseAgreementFromHtml(ctx context.Context, logger *zap.Logger, body []byte, data *model.AgreementParesedData) (any, error) {
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	data.Customer.Code, _ = getFromHtmlByTitle(doc, "td", "Идентификационный код заказчика:")
	data.Customer.Name, _ = getFromHtmlByTitle(doc, "td", "Полное наименование организации:")
	data.Customer.INN, _ = getFromHtmlByTitle(doc, "td", "ИНН/КПП:")
	data.Customer.OKOPF, _ = getFromHtmlByTitle(doc, "td", "ОКОПФ:")
	data.Customer.OKOPF = strings.ReplaceAll(data.Customer.OKOPF, "\u00A0", " ")
	data.Customer.INN = strings.Split(strings.ReplaceAll(data.Customer.INN, " ", ""), "/")[0]
	data.PurchaseMethod, _ = getFromHtmlByTitle(doc, "td", "Способ закупки:")
	data.PurchaseMethod = strings.ReplaceAll(data.PurchaseMethod, "\u00A0", " ")
	data.Subject, _ = getFromHtmlByTitle(doc, "td", "Предмет договора:")
	var columns []string
	// Parent().Next().Find(".item-information").Eq(1).Find("tr")
	doc.Find(".headerBlock").Each(func(it int, table *goquery.Selection) {
		// var columns []string
		text := strings.TrimSpace(table.Text())
		if text == "Информация о товарах, работах, услугах" {
			table.Parent().Next().Find("tr").Each(func(i int, row *goquery.Selection) {

				row.Find("th").Each(func(j int, cell *goquery.Selection) {
					columns = append(columns, cell.Text())
				})
				if i > 0 {
					service := &model.AgreementService{}
					row.Find("td").Each(func(j int, cell *goquery.Selection) {
						switch columns[j] {
						case "Наименование товаров, работ, услуг":
							arrText := strings.Split(cell.Text(), "Тип объекта закупки:")
							service.Name = arrText[0]
							if len(arrText) > 1 {
								service.TypeObject = arrText[1]
							}
						case "Классификация по ОКПД", "ОКПД(ОКДП)":
							service.OKPD = cell.Text()
						case "Классификация по ОКПД2":
							service.OKPD2 = cell.Text()
						case "Количество (Объем)":
							service.Quantity, _ = baseutils.ParsePriceToFloat(cell.Text())
						case "Единица измерения":
							service.QuantityType = cell.Text()
						case "Количество (объем), единица измерения":
							tmp := strings.Split(cell.Text(), ",")
							service.Quantity, _ = baseutils.ParsePriceToFloat(strings.TrimSpace(tmp[0]))
							if len(tmp) > 1 {
								service.QuantityType = strings.TrimSpace(tmp[1])
							}
						case "Цена за единицу":
							tmp := strings.Split(cell.Text(), ",")
							service.UnitPrice, _ = baseutils.ParsePriceToFloat(strings.TrimSpace(tmp[0]))
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
		}
	})

	return data, nil
}

func ParseCustomerFromMain(ctx context.Context, logger *zap.Logger, body []byte, data *model.AgreementParesedData) (any, error) {
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
	el, err := getElementFromHtmlByTitle(doc, "span", "Коды основного вида деятельности по ОКВЭД")
	if err == nil {
		tmp := []string{}
		if el.Children().Length() > 0 {
			el.Children().Each(func(i int, s *goquery.Selection) {
				tmp = append(tmp, strings.TrimSpace(s.Text()))
			})
			data.Customer.MainWork = strings.Join(tmp, ",")
		} else {
			data.Customer.MainWork = el.Text()
		}
	}

	data.Customer.MainWork = strings.TrimSpace(data.Customer.MainWork)
	// data.Customer.OKOPF, _ = getFromHtmlByTitle(doc, "span", "Код по ОКОПФ")
	// data.Customer.OKOPF = strings.TrimSpace(data.Customer.OKOPF)
	// logger.Debug(fmt.Sprintf("%+v\n", data))
	return data, nil
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

func getElementFromHtmlByTitle(doc *goquery.Document, tag string, title string) (*goquery.Selection, error) {
	var result *goquery.Selection = nil
	doc.Find(tag).Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == title {
			// Ищем следующий <td>
			val := s.Parent().Find(tag).Eq(s.Index() + 1)
			result = val
		}
	})
	if result != nil {
		return result, nil
	} else {
		return nil, errors.New("No element with titile" + title)
	}
}
