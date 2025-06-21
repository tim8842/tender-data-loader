package contract_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/internal/contract"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/supplier"
	"github.com/tim8842/tender-data-loader/pkg/parser"
	"github.com/tim8842/tender-data-loader/pkg/reader"
	"go.uber.org/zap"
)

func TestParseContractIds(t *testing.T) {
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

func TestParseContractFromMain(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	dir := "../../assets/test/ParseContractFromMain"
	sd, err := parser.ParseFromDateToTime("25.11.2024")
	assert.NoError(t, err)
	ed, err := parser.ParseFromDateToTime("31.12.2025")
	assert.NoError(t, err)
	ud, err := parser.ParseFromDateToTime("20.06.2025")
	assert.NoError(t, err)
	pd, err := parser.ParseFromDateToTime("26.11.2024")
	assert.NoError(t, err)
	tests := []struct {
		name      string
		inputHTML []byte
		expectErr bool
		expected  *contract.ContractParesedData
	}{
		{
			name:      "bad doc",
			inputHTML: reader.ReadHtmlFile(dir + "/error.html"),
			expectErr: true,
			expected:  nil,
		},
		{
			name:      "correct doc",
			inputHTML: reader.ReadHtmlFile(dir + "/correct.html"),
			expectErr: false,
			expected: &contract.ContractParesedData{
				ID:             "3470200936124000073",
				Status:         "Исполнение",
				NoticeId:       "0345300029324000079",
				Price:          951128.32,
				SignedAt:       sd,
				ExecutionEnd:   ed,
				UpdatedAt:      ud,
				SupplierMethod: "Электронный аукцион",
				Subject:        "Поставка продуктов. Хлеб.",
				PublishedAt:    pd,
				Customer: &customer.Customer{
					ID:   "03453000293",
					Name: `ЛЕНИНГРАДСКОЕ ОБЛАСТНОЕ ГОСУДАРСТВЕННОЕ БЮДЖЕТНОЕ УЧРЕЖДЕНИЕ "ВОЛХОВСКИЙ КОМПЛЕКСНЫЙ ЦЕНТР СОЦИАЛЬНОГО ОБСЛУЖИВАНИЯ НАСЕЛЕНИЯ "БЕРЕНИКА"`,
					INN:  "4702009361",
				}, Services: []*contract.ContractService{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := contract.ParseContractFromMain(ctx, logger, tt.inputHTML)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseContractFromHtml(t *testing.T) {
	ctx := context.Background()
	eed, err := parser.ParseFromDateToTime("08.09.2020")
	assert.NoError(t, err)
	eed2, err := parser.ParseFromDateToTime("15.01.2025")
	assert.NoError(t, err)
	logger, _ := zap.NewDevelopment()
	dir := "../../assets/test/ParseContractFromHtml"
	tests := []struct {
		name      string
		inputHTML []byte
		expectErr bool
		input     *contract.ContractParesedData
		expected  *contract.ContractParesedData
	}{
		{
			name:      "bad doc",
			inputHTML: reader.ReadHtmlFile(dir + "/error.html"),
			expectErr: true,
			expected:  nil,
		},
		{
			name:      "correct doc",
			inputHTML: reader.ReadHtmlFile(dir + "/correct2020.html"),
			expectErr: false,
			input: &contract.ContractParesedData{
				Suppliers: []*supplier.Supplier{},
				Services:  []*contract.ContractService{}},
			expected: &contract.ContractParesedData{
				ExecutionStart: eed,
				Suppliers: []*supplier.Supplier{
					{
						ID:           "6321405176632101001",
						Name:         `Общества с ограниченной ответственностью ОБЩЕСТВО С ОГРАНИЧЕННОЙ ОТВЕТСТВЕННОСТЬЮ "КОМПАНИЯ "НЕВСКИЙ"`,
						Country:      "Российская Федерация 643",
						Location:     "Российская Федерация, 445039, ОБЛ САМАРСКАЯ, Г ТОЛЬЯТТИ, Б-Р ГАЯ, ДОМ 27, КВАРТИРА 82 36740000001",
						MailLocation: "Российская Федерация, 445039, ОБЛ САМАРСКАЯ, Г ТОЛЬЯТТИ, Б-Р ГАЯ, ДОМ 27, КВАРТИРА 82",
						INN:          "6321405176",
						Contact:      "79171222581 nevskiy409545@gmail.com",
						Status:       "субъект малого предпринимательства",
					},
				},
				Services: []*contract.ContractService{
					{
						Name:            "Перчатки смотровые/процедурные из латекса гевеи, неопудренные, нестерильные",
						TypeObject:      "Товар",
						OKPD2:           "Перчатки смотровые/процедурные из латекса гевеи, неопудренные, нестерильные (22.19.60.119-00000002)",
						Quantity:        95000,
						QuantityType:    "Пара (2 шт.) (пар)",
						UnitPrice:       19.40,
						CountryOfOrigin: "МАЛАЙЗИЯ (458)",
					},
					{
						Name:            "Перчатки хирургические из латекса гевеи, неопудренные, антибактериальные",
						TypeObject:      "Товар",
						OKPD2:           "Перчатки хирургические из латекса гевеи, неопудренные, антибактериальные (22.19.60.113-00000003)",
						Quantity:        1450,
						QuantityType:    "Пара (2 шт.) (пар)",
						UnitPrice:       39.7,
						CountryOfOrigin: "Китайская Народная Республика (156)",
					},
					{
						Name:            "Перчатки хирургические из латекса гевеи, неопудренные, антибактериальные",
						TypeObject:      "Товар",
						OKPD2:           "Перчатки хирургические из латекса гевеи, неопудренные, антибактериальные (22.19.60.113-00000003)",
						Quantity:        50,
						QuantityType:    "Пара (2 шт.) (пар)",
						UnitPrice:       39.25,
						CountryOfOrigin: "Китайская Народная Республика (156)",
					},
				}},
		},
		{
			name:      "correct doc 2025",
			inputHTML: reader.ReadHtmlFile(dir + "/correct2025.html"),
			expectErr: false,
			input: &contract.ContractParesedData{
				Suppliers: []*supplier.Supplier{},
				Services:  []*contract.ContractService{}},
			expected: &contract.ContractParesedData{
				ExecutionStart: eed2,
				Suppliers: []*supplier.Supplier{
					{
						ID:           "7724053916772401001",
						Name:         `АКЦИОНЕРНОЕ ОБЩЕСТВО "ЦЕНТР ВНЕДРЕНИЯ "ПРОТЕК". АО ЦВ ПРОТЕК`,
						Country:      "Российская Федерация 643",
						Location:     "115201, Г.МОСКВА 77, Ш КАШИРСКОЕ, Д. 22, К. 4",
						MailLocation: "115201, Г.МОСКВА 77, Ш КАШИРСКОЕ, Д. 22, К. 4",
						INN:          "7724053916",
						Contact:      "7-342-2700370 a_moiseeva@perm.protek.ru",
						Status:       "субъект малого предпринимательства",
					},
					{
						ID:           "7724053916168443001",
						Name:         `Филиал АКЦИОНЕРНОГО ОБЩЕСТВА "ЦЕНТР ВНЕДРЕНИЯ "ПРОТЕК" "ПРОТЕК-10". Филиал АО "ЦВ "ПРОТЕК" "ПРОТЕК-10" Является обособленным подразделением юридического лица`,
						Country:      "Российская Федерация 643",
						Location:     "420054, РЕСПУБЛИКА ТАТАРСТАН (ТАТАРСТАН), г.о. ГОРОД КАЗАНЬ, Г КАЗАНЬ, УЛ КРУТОВСКАЯ, ЗД. 33",
						MailLocation: "115201, Г.МОСКВА 77, Ш КАШИРСКОЕ, Д. 22, К. 4",
						INN:          "7724053916",
						Contact:      "7-342-2700370 a_moiseeva@perm.protek.ru",
						Status:       "субъект малого предпринимательства",
					},
				},
				Services: []*contract.ContractService{
					{
						Name:            "Ворикоз",
						TypeObject:      "Товар",
						OKPD2:           "Препараты противогрибковые для системного использования (21.20.10.192)",
						Quantity:        14,
						QuantityType:    "Штука (шт) (таблетка)",
						UnitPrice:       519.608571428,
						CountryOfOrigin: "Республика Индия (356)",
					},
				}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := contract.ParseContractFromHtml(ctx, logger, tt.inputHTML, tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestParseCustomerFromMain(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	dir := "../../assets/test/ParseCustomerFromMain"
	tests := []struct {
		name         string
		inputHTML    []byte
		inputData    *contract.ContractParesedData
		expectErr    bool
		expectedData *contract.ContractParesedData
	}{
		{
			name:      "correctOrg",
			inputHTML: reader.ReadHtmlFile(dir + "/correctOrg.html"),
			inputData: &contract.ContractParesedData{
				Customer: &customer.Customer{},
			},
			expectErr: false,
			expectedData: &contract.ContractParesedData{
				Customer: &customer.Customer{
					Location: "Российская Федерация, 352931, Краснодарский край, Армавир г, Калинина ул, Д.32",
					OKOPF:    "75403",
					Code:     "32302032098230201001",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := contract.ParseCustomerFromMain(ctx, logger, tt.inputHTML, tt.inputData)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, tt.inputData)
			}
		})
	}
}

func TestParseCustomerFromMainAddInfo(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	dir := "../../assets/test/ParseCustomerFromMain"
	tests := []struct {
		name         string
		inputHTML    []byte
		inputData    *contract.ContractParesedData
		expectErr    bool
		expectedData *contract.ContractParesedData
	}{
		{
			name:      "correctOrgAddInfo",
			inputHTML: reader.ReadHtmlFile(dir + "/correctOrgAddInfo.html"),
			inputData: &contract.ContractParesedData{
				Customer: &customer.Customer{},
			},
			expectErr: false,
			expectedData: &contract.ContractParesedData{
				Customer: &customer.Customer{
					MainWork: "85.14: Образование среднее общее,85.41: Образование дополнительное детей и взрослых",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := contract.ParseCustomerFromMainAddInfo(ctx, logger, tt.inputHTML, tt.inputData)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, tt.inputData)
			}
		})
	}
}
