package contract

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/supplier"
	"github.com/tim8842/tender-data-loader/pkg/parser"
	"go.uber.org/zap"
)

func ParseContractIds(ctx context.Context, logger *zap.Logger, data []byte) (any, error) {
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
		id := parser.GetParamFromHref(href, "reestrNumber")

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

func ParseContractFromMain(ctx context.Context, logger *zap.Logger, body []byte) (any, error) {
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, errors.New("failed to parse HTML: " + error.Error(err))
	}
	data := &ContractParesedData{Customer: &customer.Customer{}, Services: []*ContractService{}}
	noticeSpan, err := parser.GetElementFromHtmlByTitle(doc, "span", "Номер извещения об осуществлении закупки")
	if err == nil {
		noticeSpan.Children().Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				data.NoticeId = parser.GetParamFromHref(href, "regNumber")
			}

		})
	}
	doc.Find("span.cardMainInfo__purchaseLink.distancedText a").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		text = strings.TrimPrefix(text, "№")
		data.ID = strings.ReplaceAll(text, " ", "")
	})
	if data.NoticeId == "" {
		logger.Debug("Error contract have no noticeId " + data.ID)
		return nil, errors.New("error contract have no noticeId")
	}
	// 1. Номер (№ 82902040527250000050000 → 82902040527250000050000
	// 2. Статус
	doc.Find("span.cardMainInfo__state").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		data.Status = text
	})
	// 2. Цена
	doc.Find(".cardMainInfo__content.cost").EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		text = strings.ReplaceAll(text, "₽", "")
		text = strings.ReplaceAll(text, ",", ".")
		if price, err := parser.ParsePriceToFloat(text); err == nil {
			data.Price = price
			return false // найдена первая цена
		}
		return true
	})
	tmp, _ := parser.GetFromHtmlByTitle(doc, "span", "Заключение контракта")
	data.Customer.Name, _ = parser.GetFromHtmlByTitle(doc, "span", "Полное наименование заказчика")
	data.Customer.Name = strings.TrimSpace(data.Customer.Name)
	data.Customer.INN, _ = parser.GetFromHtmlByTitle(doc, "span", "ИНН")
	data.SignedAt, _ = parser.ParseFromDateToTime(strings.TrimSpace(tmp))
	tmp, _ = parser.GetFromHtmlByTitle(doc, "span", "Размещен контракт в реестре контрактов")
	data.PublishedAt, _ = parser.ParseFromDateToTime(strings.TrimSpace(tmp))
	tmp, _ = parser.GetFromHtmlByTitle(doc, "span", "Срок исполнения")
	data.ExecutionEnd, _ = parser.ParseFromDateToTime(strings.TrimSpace(tmp))
	tmp, _ = parser.GetFromHtmlByTitle(doc, "span", "Обновлен контракт в реестре контрактов")
	data.UpdatedAt, _ = parser.ParseFromDateToTime(strings.TrimSpace(tmp))
	// 6. agencyId из 9-го .section__info a
	tmp, _ = parser.GetFromHtmlByTitle(doc, "span", "Способ определения поставщика (подрядчика, исполнителя)")
	data.SupplierMethod = strings.TrimSpace(tmp)
	tmp, _ = parser.GetFromHtmlByTitle(doc, "span", "Предмет контракта")
	data.Subject = strings.TrimSpace(tmp)
	agencyIdSpan, err := parser.GetElementFromHtmlByTitle(doc, "span", "Заказчик")
	if err == nil {
		agencyIdSpan.Find("a").Each(func(i int, s *goquery.Selection) {
			if href, ok := s.Attr("href"); ok {
				data.Customer.ID = parser.GetParamFromHref(href, "organizationCode")
			}

		})
	}
	if data.Customer.ID == "" {
		logger.Debug("Error no ID from customer")
		err = errors.New("error no customer id contract")
	}
	return data, err
}

func RemoveBr(ss *goquery.Selection, instead string) {
	ss.Find("br").Each(func(i int, s *goquery.Selection) {
		// Вставляем текстовый узел со строкой " " перед <br>
		s.BeforeHtml(instead)
		// Удаляем сам <br>
		s.Remove()
	})
}

func ClearTextFromTrash(text string) string {
	spaceRe := regexp.MustCompile(`\s+`)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\u00a0", " ")
	text = spaceRe.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	return text
}

func ParseContractFromHtml(ctx context.Context, logger *zap.Logger, body []byte, data *ContractParesedData) (res any, err error) {
	defer func() {
		if r := recover(); r != nil {
			var id string
			if data != nil {
				id = data.ID
			}
			err = fmt.Errorf("panic caught during parsing: %v", r)
			logger.Debug("Panic caught during parsing contract html " + id)
		}
	}()
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}
	var columns []string
	// Parent().Next().Find(".item-information").Eq(1).Find("tr")
	doc.Find(".bottom-pad").Each(func(it int, table *goquery.Selection) {
		// var columns []string
		text := strings.TrimSpace(table.Text())
		spaceRe := regexp.MustCompile(`\s+`)
		if strings.Contains(text, "Объект закупки") {
			table.Next().Find("tr").Each(func(i int, row *goquery.Selection) {
				if i == 0 {
					row.Find("td").Each(func(j int, cell *goquery.Selection) {
						RemoveBr(cell, " ")
						columns = append(columns, spaceRe.ReplaceAllString(cell.Text(), " "))
					})
				}
				if i > 0 {
					service := &ContractService{}
					row.Find("td").Each(func(j int, cell *goquery.Selection) {
						RemoveBr(cell, " ")
						switch columns[j] {
						case "Наименование объекта закупки":
							service.Name = ClearTextFromTrash(cell.Text())
						case "Тип объекта закупки":
							service.TypeObject = ClearTextFromTrash(cell.Text())
						case "Классификация по ОКПД", "ОКПД(ОКДП)":
							service.OKPD = ClearTextFromTrash(cell.Text())
						case "Позиции по КТРУ, ОКПД2, информация о ТРУ", "Код по КТРУ, ОКПД2":
							service.OKPD2 = ClearTextFromTrash(cell.Text())
						case "Количество (Объем)", "Количество":
							service.Quantity, _ = parser.ParsePriceToFloat(cell.Text())
						case "Единица измерения", "Единица измерения товара, работы, услуги":
							service.QuantityType = ClearTextFromTrash(cell.Text())
						case "Количество (объем) и единица измерения по ОКЕИ", "Количество и единица измерения по ОКЕИ":
							tmp := strings.SplitN(cell.Text(), " ", 2)
							service.Quantity, _ = parser.ParsePriceToFloat(strings.TrimSpace(tmp[0]))
							if len(tmp) > 1 {
								service.QuantityType = strings.TrimSpace(spaceRe.ReplaceAllString(tmp[1], " "))
							}
						case "Цена за единицу (в валюте контракта)":
							service.UnitPrice, _ = parser.ParsePriceToFloat(strings.TrimSpace(cell.Text()))
						case "Страна происхождения товара", "Страна происхождения":
							service.CountryOfOrigin = ClearTextFromTrash(cell.Text())
						}
					})
					data.Services = append(data.Services, service)
				}
			})
		} else if strings.Contains(text, "Информация о поставщиках") {
			columns = []string{}
			var columns1 = []string{}
			var columns2 = []string{}
			needBiggerThan := 1
			table.Next().Find("tr").Each(func(i int, row *goquery.Selection) {
				if i == 0 {
					row.Find("td").Each(func(j int, cell *goquery.Selection) {
						RemoveBr(cell, " ")
						columns1 = append(columns1, spaceRe.ReplaceAllString(cell.Text(), " "))
					})
				}
				if !slices.Contains(columns1, "Адрес места нахождения (места жительства)") {
					if i == 1 {
						columns = []string{}
						row.Find("td").Each(func(j int, cell *goquery.Selection) {
							RemoveBr(cell, " ")
							columns2 = append(columns2, spaceRe.ReplaceAllString(cell.Text(), " "))
						})
						columns = append(columns, columns1[0:2]...)
						columns = append(columns, columns2...)
						columns = append(columns, columns1[4:]...)

					}
					needBiggerThan = 1
				} else {
					needBiggerThan = 0
					columns = columns1
				}
				if i > needBiggerThan {
					supplier := &supplier.Supplier{}
					var kpp string = ""
					row.Find("td").Each(func(j int, cell *goquery.Selection) {
						RemoveBr(cell, " ")
						switch columns[j] {
						case "Наименование юридического лица (ф.и.о. физического лица)", "Наименование юридического лица (Ф.И.О. физического лица)":
							supplier.Name = ClearTextFromTrash(cell.Text())
						case "Наименование страны, код по ОКСМ":
							supplier.Country = ClearTextFromTrash(cell.Text())
						case "Адрес в стране регистрации (для иностранных поставщиков)":
							if cell.Text() != "" {
								supplier.Location = ClearTextFromTrash(cell.Text())
							}
						case "Адрес, код по ОКТМО", "Адрес места нахождения (места жительства)":
							if cell.Text() != "" {
								supplier.Location = ClearTextFromTrash(cell.Text())
							}
						case "Адрес пользователя услугами почтовой связи", "Почтовый адрес":
							supplier.MailLocation = ClearTextFromTrash(cell.Text())
						case "КПП, дата постановки на учет":
							kpp = ClearTextFromTrash(cell.Text())
							kpp = strings.Split(kpp, ",")[0]
						case "КПП":
							kpp = ClearTextFromTrash(cell.Text())
						case "ИНН":
							supplier.INN = ClearTextFromTrash(cell.Text())
							if supplier.INN == "" {
								logger.Debug("Error no inn customer")
								err = errors.New("error no inn")
								return
							}
						case "Телефон (электронная почта)":
							supplier.Contact = ClearTextFromTrash(cell.Text())
						case "Статус":
							supplier.Status = ClearTextFromTrash(cell.Text())
						}

					})
					if kpp == "" {
						logger.Debug("no kpp supplier " + data.ID)
					}
					if supplier.INN == "" {
						logger.Debug("no INN supplier " + data.ID)
					}
					supplier.ID = supplier.INN + kpp
					if supplier.Name != "" && supplier.INN != "" { //&& kpp != ""
						data.Suppliers = append(data.Suppliers, supplier)
					}
				}
			})
		}

	})
	// fmt.Println(data.ID)
	// fmt.Println(len(data.Services))
	// fmt.Println(len(data.Suppliers))
	data.Services = data.Services[1 : len(data.Services)-1]
	data.Suppliers = data.Suppliers[1:len(data.Suppliers)]
	tmp, _ := parser.GetFromHtmlByTitle(doc, "td", "Дата начала исполнения контракта")
	tmpTime, _ := parser.ParseFromDateToTime(tmp)
	data.ExecutionStart = tmpTime
	if len(data.Suppliers) == 0 {
		logger.Debug("error no supplier " + data.ID)
		return nil, errors.New("error no supplier")
	}
	return data, err
}

func ParseCustomerFromMain(ctx context.Context, logger *zap.Logger, body []byte, data *ContractParesedData) (any, error) {
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

	data.Customer.OKOPF, _ = parser.GetFromHtmlByTitle(doc, "span", "Код по ОКОПФ")
	data.Customer.OKOPF = strings.TrimSpace(data.Customer.OKOPF)
	data.Customer.Code, _ = parser.GetFromHtmlByTitle(doc, "span", "ИКУ")
	data.Customer.Code = strings.TrimSpace(data.Customer.Code)
	// logger.Debug(fmt.Sprintf("%+v\n", data))
	return data, nil
}

func ParseCustomerFromMainAddInfo(ctx context.Context, logger *zap.Logger, body []byte, data *ContractParesedData) (any, error) {
	reader := bytes.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	el, err := parser.GetElementFromHtmlByTitle(doc, "span", "ОКВЭД")
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

	return data, nil
}
