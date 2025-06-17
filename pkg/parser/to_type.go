package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func ParsePriceToFloat(input string) (float64, error) {
	// Удаляем пробелы и неразрывные пробелы
	cleaned := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\u00A0' {
			return -1
		}
		if unicode.IsDigit(r) || r == ',' || r == '.' {
			return r
		}
		return -1
	}, input)

	// Определяем последнюю запятую или точку как десятичный разделитель
	lastComma := strings.LastIndex(cleaned, ",")
	lastDot := strings.LastIndex(cleaned, ".")

	var decimalIndex int
	var decimalRune rune

	switch {
	case lastComma > lastDot:
		decimalIndex = lastComma
		decimalRune = ','
	case lastDot > lastComma:
		decimalIndex = lastDot
		decimalRune = '.'
	default:
		decimalIndex = -1 // нет десятичного разделителя
	}

	var builder strings.Builder
	for i, r := range cleaned {
		if (r == ',' || r == '.') && i != decimalIndex {
			continue // удалить как разделитель тысяч
		}
		builder.WriteRune(r)
	}

	// Заменяем десятичный разделитель на точку для ParseFloat
	normalized := strings.ReplaceAll(builder.String(), string(decimalRune), ".")

	return strconv.ParseFloat(normalized, 64)
}

func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func ParseFromDateToTime(dateString string) (time.Time, error) {
	t, err := time.Parse("02.01.2006", dateString)
	if err != nil {
		return time.Time{}, fmt.Errorf("ошибка при парсинге даты: %w", err)
	}
	return t, nil
}

func FromTimeToDate(t time.Time) string {
	return t.Format("02.01.2006")
}
