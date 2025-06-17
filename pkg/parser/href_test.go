package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/pkg/parser"
)

func TestGetParamFromHref(t *testing.T) {
	tests := []struct {
		name     string
		href     string
		param    string
		expected string
	}{
		{
			name:     "without params",
			href:     "users/gello",
			param:    "id",
			expected: "",
		},
		{
			name:     "two different params",
			href:     "users/gello?id=132&cat=green",
			param:    "cat",
			expected: "green",
		},
		{
			name:     "no param",
			href:     "users/gello?id=132",
			param:    "cat",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.GetParamFromHref(tt.href, tt.param)
			assert.Equal(t, tt.expected, result)
		})
	}
}
