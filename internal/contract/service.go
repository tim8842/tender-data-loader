package contract

import (
	"bytes"
	"context"

	"github.com/PuerkitoBio/goquery"
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
