package variable_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/internal/variable"
)

func TestVariableBackToNowAgreement_ConvertToVariable(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name     string
		input    variable.VariableBackToNowAgreement
		wantErr  bool
		expected variable.Variable
	}{
		{
			name: "valid conversion",
			input: variable.VariableBackToNowAgreement{
				ID: "abc123",
				Vars: variable.VarsBackToNowAgreement{
					Page:     2,
					SignedAt: now,
				},
			},
			wantErr: false,
			expected: variable.Variable{
				ID: "abc123",
				Vars: map[string]interface{}{
					"page":      float64(2), // JSON numbers are float64 by default
					"signed_at": now.Format(time.RFC3339),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.ConvertToVariable()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Проверяем ID
				assert.Equal(t, tt.expected.ID, result.ID)

				// Проверяем, что page верно преобразован
				assert.Equal(t, tt.expected.Vars["page"], result.Vars["page"])

				// Проверяем, что signed_at верно сериализован
				signedAtStr, ok := result.Vars["signed_at"].(string)
				assert.True(t, ok)
				parsedTime, err := time.Parse(time.RFC3339, signedAtStr)
				assert.NoError(t, err)
				assert.True(t, parsedTime.Equal(now))
			}
		})
	}
}
