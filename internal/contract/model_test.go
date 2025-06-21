package contract_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/internal/contract"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/supplier"
)

func TestParseContractParsedDataToModel(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name              string
		input             *contract.ContractParesedData
		expected          *contract.Contract
		expectedCustomer  *customer.Customer
		expectedSuppliers []*supplier.Supplier
	}{
		{
			name: "basic mapping with one supplier",
			input: &contract.ContractParesedData{
				ID:             "contract123",
				Status:         "active",
				NoticeId:       "notice001",
				Price:          10000,
				SignedAt:       now,
				ExecutionStart: now,
				ExecutionEnd:   now.AddDate(0, 1, 0),
				PublishedAt:    now,
				UpdatedAt:      now,
				Law:            "44-ФЗ",
				SupplierMethod: "auction",
				Subject:        "office supplies",
				Suppliers: []*supplier.Supplier{
					{ID: "supplier1"},
				},
				Customer: &customer.Customer{
					ID:       "customer1",
					Code:     "001",
					Name:     "Test Customer",
					INN:      "1234567890",
					OKOPF:    "12345",
					MainWork: "Gov services",
					Location: "Moscow",
				},
				Services: []*contract.ContractService{
					{
						Name:         "Paper A4",
						TypeObject:   "goods",
						Quantity:     100,
						QuantityType: "pcs",
						UnitPrice:    10,
						Currency:     "RUB",
					},
				},
			},
			expected: &contract.Contract{
				ID:             "contract123",
				Status:         "active",
				NoticeId:       "notice001",
				Price:          10000,
				SignedAt:       now,
				ExecutionStart: now,
				ExecutionEnd:   now.AddDate(0, 1, 0),
				PublishedAt:    now,
				UpdatedAt:      now,
				Law:            "44-ФЗ",
				SupplierMethod: "auction",
				Subject:        "office supplies",
				SupplierIds:    []string{"supplier1"},
				CustomerId:     "customer1",
				Services: []*contract.ContractService{
					{
						Name:         "Paper A4",
						TypeObject:   "goods",
						Quantity:     100,
						QuantityType: "pcs",
						UnitPrice:    10,
						Currency:     "RUB",
					},
				},
			},
			expectedCustomer: &customer.Customer{
				ID:       "customer1",
				Code:     "001",
				Name:     "Test Customer",
				INN:      "1234567890",
				OKOPF:    "12345",
				MainWork: "Gov services",
				Location: "Moscow",
			},
			expectedSuppliers: []*supplier.Supplier{
				{ID: "supplier1"},
			},
		},
		{
			name: "no suppliers",
			input: &contract.ContractParesedData{
				ID:       "contract124",
				Price:    5000,
				Customer: &customer.Customer{ID: "cust2"},
			},
			expected: &contract.Contract{
				ID:          "contract124",
				Price:       5000,
				SupplierIds: []string{},
				CustomerId:  "cust2",
			},
			expectedCustomer: &customer.Customer{
				ID: "cust2",
			},
			expectedSuppliers: []*supplier.Supplier{},
		},
		{
			name: "nil supplier skipped",
			input: &contract.ContractParesedData{
				ID: "contract125",
				Suppliers: []*supplier.Supplier{
					nil,
					{ID: ""},
					{ID: "supplier2"},
				},
				Customer: &customer.Customer{ID: "cust3"},
			},
			expected: &contract.Contract{
				ID:          "contract125",
				SupplierIds: []string{"supplier2"},
				CustomerId:  "cust3",
			},
			expectedCustomer: &customer.Customer{
				ID: "cust3",
			},
			expectedSuppliers: []*supplier.Supplier{
				{ID: "supplier2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contractModel, cust, suppliers := contract.ParseContractParsedDataToModel(tt.input)

			assert.Equal(t, tt.expected, contractModel)
			assert.Equal(t, tt.expectedCustomer, cust)
			assert.Equal(t, tt.expectedSuppliers, suppliers)
		})
	}
}
