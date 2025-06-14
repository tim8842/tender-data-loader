package baseutils

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func FormatDate(t time.Time) string {
	return t.Format("02.01.2006")
}

func DateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
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

func TruncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return s
}

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

func ReadHtmlFile(path string) []byte {
	if data, err := os.ReadFile(path); err == nil {
		return data
	}
	return nil
}

// func ConvertStruct(dst, src interface{}) error {
// 	srcVal := reflect.ValueOf(src)
// 	dstVal := reflect.ValueOf(dst)

// 	if srcVal.Kind() == reflect.Ptr {
// 		srcVal = srcVal.Elem()
// 	}
// 	if dstVal.Kind() != reflect.Ptr || dstVal.Elem().Kind() != reflect.Struct {
// 		return fmt.Errorf("dst must be pointer to struct")
// 	}
// 	dstVal = dstVal.Elem()

// 	if srcVal.Kind() != reflect.Struct {
// 		return fmt.Errorf("src must be struct or pointer to struct")
// 	}

// 	for i := 0; i < srcVal.NumField(); i++ {
// 		srcField := srcVal.Field(i)
// 		srcFieldType := srcVal.Type().Field(i)

// 		dstField := dstVal.FieldByName(srcFieldType.Name)
// 		if !dstField.IsValid() || !dstField.CanSet() {
// 			continue
// 		}

// 		err := setValue(dstField, srcField)
// 		if err != nil {
// 			return fmt.Errorf("field %s: %w", srcFieldType.Name, err)
// 		}
// 	}
// 	return nil
// }

// func setValue(dst, src reflect.Value) error {
// 	// Если это указатели, распаковываем
// 	if src.Kind() == reflect.Ptr {
// 		if src.IsNil() {
// 			return nil
// 		}
// 		src = src.Elem()
// 	}
// 	if dst.Kind() == reflect.Ptr {
// 		if dst.IsNil() {
// 			dst.Set(reflect.New(dst.Type().Elem()))
// 		}
// 		dst = dst.Elem()
// 	}

// 	// Если оба структуры — рекурсивно вызываем
// 	if src.Kind() == reflect.Struct && dst.Kind() == reflect.Struct {
// 		return ConvertStruct(dst.Addr().Interface(), src.Interface())
// 	}

// 	// Попытка конвертации базовых типов
// 	switch dst.Kind() {
// 	case reflect.String:
// 		switch src.Kind() {
// 		case reflect.String:
// 			dst.SetString(src.String())
// 		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 			dst.SetString(strconv.FormatInt(src.Int(), 10))
// 		case reflect.Float32, reflect.Float64:
// 			dst.SetString(strconv.FormatFloat(src.Float(), 'f', -1, 64))
// 		default:
// 			return fmt.Errorf("cannot convert %s to string", src.Kind())
// 		}
// 		return nil
// 	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 		switch src.Kind() {
// 		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
// 			dst.SetInt(src.Int())
// 		case reflect.String:
// 			iv, err := strconv.ParseInt(src.String(), 10, 64)
// 			if err != nil {
// 				return err
// 			}
// 			dst.SetInt(iv)
// 		default:
// 			return fmt.Errorf("cannot convert %s to int", src.Kind())
// 		}
// 		return nil
// 	case reflect.Float32, reflect.Float64:
// 		switch src.Kind() {
// 		case reflect.Float32, reflect.Float64:
// 			dst.SetFloat(src.Float())
// 		case reflect.String:
// 			fv, err := strconv.ParseFloat(src.String(), 64)
// 			if err != nil {
// 				return err
// 			}
// 			dst.SetFloat(fv)
// 		default:
// 			return fmt.Errorf("cannot convert %s to float", src.Kind())
// 		}
// 		return nil
// 	}

// 	// Если типы совпадают и присваиваемы — просто копируем
// 	if src.Type().AssignableTo(dst.Type()) {
// 		dst.Set(src)
// 		return nil
// 	}

// 	return fmt.Errorf("unsupported conversion from %s to %s", src.Type(), dst.Type())
// }
