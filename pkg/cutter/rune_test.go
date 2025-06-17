package cutter_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tim8842/tender-data-loader/pkg/cutter"
)

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
			require.Equal(t, tt.expected, cutter.TruncateRunes(tt.input, tt.maxLen))
		})
	}
}
