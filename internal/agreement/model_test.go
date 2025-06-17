package agreement_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/customer"
)

func TestParseAgreementDataToModels(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    *agreement.AgreementParesedData
		expected *agreement.Agreement
	}{
		{
			name: "basic agreement and customer",
			input: &agreement.AgreementParesedData{
				ID:             "agr123",
				Number:         "A-001",
				Status:         "active",
				Pfid:           "pfid123",
				NoticeId:       "not456",
				Price:          150000,
				SignedAt:       now,
				ExecutionStart: now,
				ExecutionEnd:   now,
				PublishedAt:    now,
				UpdatedAt:      now,
				PurchaseMethod: "auction",
				Subject:        "IT services",
				Customer: &customer.Customer{
					ID:       "cust001",
					Code:     "12345678",
					Name:     "Test Customer",
					INN:      "1234567890",
					OKOPF:    "12300",
					MainWork: "Software development",
					Location: "Moscow",
				},
				Services: []*agreement.AgreementService{
					{
						Name:         "Development",
						UnitPrice:    50000,
						Quantity:     3,
						Currency:     "RUB",
						QuantityType: "unit",
					},
				},
			},
			expected: &agreement.Agreement{
				ID:             "agr123",
				Number:         "A-001",
				Status:         "active",
				Pfid:           "pfid123",
				NoticeId:       "not456",
				Price:          150000,
				SignedAt:       now,
				ExecutionStart: now,
				ExecutionEnd:   now,
				PublishedAt:    now,
				UpdatedAt:      now,
				PurchaseMethod: "auction",
				Subject:        "IT services",
				CustomerId:     "cust001",
				Services: []*agreement.AgreementService{
					{
						Name:         "Development",
						UnitPrice:    50000,
						Quantity:     3,
						Currency:     "RUB",
						QuantityType: "unit",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agreement, customer := agreement.ParseAgreementDataToModels(tt.input)
			assert.Equal(t, tt.expected, agreement)
			assert.Equal(t, tt.input.Customer.ID, customer.ID)
			assert.Equal(t, tt.input.Customer.Name, customer.Name)
		})
	}
}
