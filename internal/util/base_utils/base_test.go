package baseutils

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// --- FormatDate ---
func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{"normal date", time.Date(2024, 12, 25, 15, 4, 5, 0, time.UTC), "25.12.2024"},
		{"beginning of year", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "01.01.2025"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, FormatDate(tt.input))
		})
	}
}

// --- DateOnly ---
func TestDateOnly(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected time.Time
	}{
		{"truncate time", time.Date(2025, 6, 12, 14, 30, 45, 1234, time.UTC), time.Date(2025, 6, 12, 0, 0, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, DateOnly(tt.input))
		})
	}
}

// --- ParseDate ---
func TestParseDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  time.Time
		expectErr bool
	}{
		{"valid date", "15.04.2024", time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC), false},
		{"invalid format", "2024-04-15", time.Time{}, true},
		{"empty string", "", time.Time{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseDate(tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// --- UnmarshalVars ---
func TestUnmarshalVars(t *testing.T) {
	type Sample struct {
		Name string
		Age  int
	}
	tests := []struct {
		name      string
		input     map[string]interface{}
		expected  Sample
		expectErr bool
	}{
		{
			"valid input",
			map[string]interface{}{"Name": "Alice", "Age": 30},
			Sample{Name: "Alice", Age: 30},
			false,
		},
		{
			"invalid type",
			map[string]interface{}{"Name": "Bob", "Age": "not a number"},
			Sample{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := UnmarshalVars[Sample](tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}

// --- TruncateRunes ---
func TestTruncateRunes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"shorter", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncate", "hello world", 5, "hello"},
		{"unicode", "привет", 4, "прив"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, TruncateRunes(tt.input, tt.maxLen))
		})
	}
}

// --- ParsePriceToFloat ---
func TestParsePriceToFloat(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  float64
		expectErr bool
	}{
		{"simple", "123.45", 123.45, false},
		{"comma as dot", "123,45", 123.45, false},
		{"with spaces", " 1 234,56 ", 1234.56, false},
		{"with nbsp", "1 234,56", 1234.56, false}, // \u00A0
		{"letters inside", "USD 1,234.56", 1234.56, false},
		{"invalid input", "abc", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePriceToFloat(tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.InDelta(t, tt.expected, result, 0.01)
			}
		})
	}
}

// --- ReadHtmlFile (интеграционный тест с временным файлом) ---
func TestReadHtmlFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("test_docs", "test.html")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := "<html><body>Test</body></html>"
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	result := ReadHtmlFile(tmpFile.Name())
	require.Equal(t, []byte(content), result)
}
