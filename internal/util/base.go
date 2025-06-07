package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func FormatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

func ParseDate(dateString string) (time.Time, error) {
	t, err := time.Parse("02.01.2006", dateString)
	if err != nil {
		return time.Time{}, fmt.Errorf("ошибка при парсинге даты: %w", err)
	}
	return t, nil
}

func UnmarshalVars[T any](vars map[string]interface{}) (T, error) {
	var result T
	bytes, err := json.Marshal(vars)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(bytes, &result)
	return result, err
}

func truncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}

func ParsePriceToFloat(input string) (float64, error) {
	// Удаляем пробелы, неразрывные пробелы и все лишние символы
	cleaned := strings.Map(func(r rune) rune {
		if r == ' ' || r == '\u00A0' {
			return -1 // удалить
		}
		if unicode.IsDigit(r) || r == ',' || r == '.' {
			return r // оставить
		}
		return -1 // всё остальное удалить
	}, input)

	// Заменяем запятую на точку (если это десятичный разделитель)
	cleaned = strings.ReplaceAll(cleaned, ",", ".")

	return strconv.ParseFloat(cleaned, 64)
}
