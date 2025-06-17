package parser_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tim8842/tender-data-loader/pkg/parser"
)

// --- FormatDate ---
func TestFromTimeToDate(t *testing.T) {
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
			require.Equal(t, tt.expected, parser.FromTimeToDate(tt.input))
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
			require.Equal(t, tt.expected, parser.DateOnly(tt.input))
		})
	}
}

// --- ParseDate ---
func TestParseFromDateToTime(t *testing.T) {
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
			result, err := parser.ParseFromDateToTime(tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
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
			result, err := parser.ParsePriceToFloat(tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.InDelta(t, tt.expected, result, 0.01)
			}
		})
	}
}
