package contract_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/internal/contract"
	"github.com/tim8842/tender-data-loader/pkg/reader"
	"go.uber.org/zap"
)

func TestParseAgreementIds(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	dir := "../../assets/test/ParseContractIds"

	tests := []struct {
		name      string
		inputHTML []byte
		expectErr bool
		expected  []string
	}{
		{
			name:      "bad doc",
			inputHTML: reader.ReadHtmlFile(dir + "/error.html"),
			expectErr: false,
			expected:  []string{},
		},
		{
			name:      "correct doc",
			inputHTML: reader.ReadHtmlFile(dir + "/correct.html"),
			expectErr: false,
			expected: []string{
				"2420538736325000069", "3366303601425000009", "3744703300925000006", "3070200700725000002", "1380818408024000013",
				"3421703831025000007", "3510302046324000007", "2562300501324000224", "2160200163824000056", "2781701531024000374",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := contract.ParseContractIds(ctx, logger, tt.inputHTML)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
