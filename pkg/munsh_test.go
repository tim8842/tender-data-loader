package pkg_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tim8842/tender-data-loader/pkg"
)

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
			result, err := pkg.UnmarshalVars[Sample](tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
		})
	}
}
