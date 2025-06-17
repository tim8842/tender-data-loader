package parser

import (
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetFromHtmlByTitle(doc *goquery.Document, tag string, title string) (string, error) {
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

func GetElementFromHtmlByTitle(doc *goquery.Document, tag string, title string) (*goquery.Selection, error) {
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
