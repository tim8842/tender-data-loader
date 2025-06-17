package agreement_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tim8842/tender-data-loader/internal/agreement"
	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/pkg/parser"
	"github.com/tim8842/tender-data-loader/pkg/reader"
	"go.uber.org/zap"
)

func TestParseAgreementIds(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	dir := "../../assets/test/ParseAgreementIds"

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
				"22455192", "20042235", "21709143", "22336931", "21063186",
				"20927331", "21980549", "22483131", "22393481", "21059010",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := agreement.ParseAgreementIds(ctx, logger, tt.inputHTML)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func getDate(data string) time.Time {
	d, _ := parser.ParseFromDateToTime(data)
	return d
}

func TestParseAgreementFromMain(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	dir := "../../assets/test/ParseAgreementFromMain"
	tests := []struct {
		name         string
		inputHTML    []byte
		expectErr    bool
		expectedData any
	}{
		{
			name:      "correct",
			inputHTML: reader.ReadHtmlFile(dir + "/correct.html"),
			expectErr: false,
			expectedData: &agreement.AgreementParesedData{
				ID: "", Number: "56680005928250002570000",
				Status: "Исполнение завершено", Pfid: "67051964",
				NoticeId: "18245882", Price: 129307.50,
				SignedAt: getDate("23.05.2025"), ExecutionStart: getDate("23.05.2025"),
				ExecutionEnd: getDate("19.06.2025"), PublishedAt: getDate("23.05.2025"),
				UpdatedAt: getDate("11.06.2025"), Customer: &customer.Customer{ID: "492275"},
			},
		},
		{
			name:         "error",
			inputHTML:    reader.ReadHtmlFile(dir + "/error.html"),
			expectErr:    false,
			expectedData: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := agreement.ParseAgreementFromMain(ctx, logger, tt.inputHTML)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
			}
		})
	}
}

func TestParseAgreementFromHtml(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	dir := "../../assets/test/ParseAgreementFromHtml"
	tests := []struct {
		name         string
		inputHTML    []byte
		inputData    *agreement.AgreementParesedData
		expectErr    bool
		expectedData *agreement.AgreementParesedData
	}{
		{
			name:      "correctNew2024",
			inputHTML: reader.ReadHtmlFile(dir + "/correctNew2024.html"),
			inputData: &agreement.AgreementParesedData{
				ID: "", Number: "56680005928250002570000",
				Status: "Исполнение завершено", Pfid: "67051964",
				NoticeId: "32514792982", Price: 129307.50,
				SignedAt: getDate("23.05.2025"), ExecutionStart: getDate("23.05.2025"),
				ExecutionEnd: getDate("19.06.2025"), PublishedAt: getDate("23.05.2025"),
				UpdatedAt: getDate("11.06.2025"), Customer: &customer.Customer{ID: "492275"},
			},
			expectErr: false,
			expectedData: &agreement.AgreementParesedData{
				ID: "", Number: "56680005928250002570000",
				Status: "Исполнение завершено", Pfid: "67051964",
				NoticeId: "32514792982", Price: 129307.50,
				SignedAt: getDate("23.05.2025"), ExecutionStart: getDate("23.05.2025"),
				ExecutionEnd: getDate("19.06.2025"), PublishedAt: getDate("23.05.2025"),
				UpdatedAt: getDate("11.06.2025"), Customer: &customer.Customer{
					ID: "492275", Code: "57705013033770501001",
					Name:  `ГОСУДАРСТВЕННОЕ УНИТАРНОЕ ПРЕДПРИЯТИЕ ГОРОДА МОСКВЫ ПО ЭКСПЛУАТАЦИИ МОСКОВСКИХ ВОДООТВОДЯЩИХ СИСТЕМ "МОСВОДОСТОК"`,
					INN:   "7705013033",
					OKOPF: "65242 Государственные унитарные предприятия субъектов Российской Федерации",
				},
				PurchaseMethod: "30000 Закупка у единственного поставщика (подрядчика, исполнителя)",
				Subject:        "Поставка инструмента аварийно-спасательного для нужд ГУП «Мосводосток»",
				Services: []*agreement.AgreementService{
					{
						Name:              "Инструмент многофункциональный аварийно-спасательный",
						TypeObject:        "Товар",
						Quantity:          1,
						QuantityType:      "Штука",
						UnitPrice:         295990.20,
						Currency:          "Российский рубль",
						CountryOfOrigin:   "Российская Федерация",
						CountryRegistered: "—",
						OKPD2:             "ОКПД2:28.24.12.190 Инструменты ручные прочие с механизированным приводом, не включенные в другие группировки",
					},
					{
						Name:              "Инструмент многофункциональный аварийно-спасательный",
						TypeObject:        "Товар",
						Quantity:          1,
						QuantityType:      "Штука",
						UnitPrice:         329373,
						Currency:          "Российский рубль",
						CountryOfOrigin:   "Российская Федерация",
						CountryRegistered: "—",
						OKPD2:             "ОКПД2:28.24.12.190 Инструменты ручные прочие с механизированным приводом, не включенные в другие группировки",
					},
					{
						Name:              "Инструмент многофункциональный аварийно-спасательный",
						TypeObject:        "Товар",
						Quantity:          1,
						QuantityType:      "Штука",
						UnitPrice:         316512.90,
						Currency:          "Российский рубль",
						CountryOfOrigin:   "Российская Федерация",
						CountryRegistered: "—",
						OKPD2:             "ОКПД2:28.24.12.190 Инструменты ручные прочие с механизированным приводом, не включенные в другие группировки",
					},
					{
						Name:              "Станок ручной для пожарных рукавов",
						TypeObject:        "Товар",
						Quantity:          1,
						QuantityType:      "Комплект",
						UnitPrice:         454622.84,
						Currency:          "Российский рубль",
						CountryOfOrigin:   "Российская Федерация",
						CountryRegistered: "—",
						OKPD2:             "ОКПД2:28.99.39.190 Оборудование специального назначения прочее, не включенное в другие группировки",
					},
				},
			},
		},
		{
			name:      "correctOld2013",
			inputHTML: reader.ReadHtmlFile(dir + "/correctOld2013.html"),
			inputData: &agreement.AgreementParesedData{
				ID: "", Number: "55110001373150000040000",
				Status: "Исполнение прекращено", Pfid: "3067",
				NoticeId: "", Price: 708602.24,
				SignedAt: getDate("10.06.2013"), ExecutionStart: getDate("10.06.2013"),
				ExecutionEnd: getDate("31.01.2016"), PublishedAt: getDate("16.01.2015"),
				UpdatedAt: getDate("02.03.2015"), Customer: &customer.Customer{ID: "87730"},
			},
			expectErr: false,
			expectedData: &agreement.AgreementParesedData{
				ID: "", Number: "55110001373150000040000",
				Status: "Исполнение прекращено", Pfid: "3067",
				NoticeId: "", Price: 708602.24,
				SignedAt: getDate("10.06.2013"), ExecutionStart: getDate("10.06.2013"),
				ExecutionEnd: getDate("31.01.2016"), PublishedAt: getDate("16.01.2015"),
				UpdatedAt: getDate("02.03.2015"), Customer: &customer.Customer{
					ID: "87730", Code: "",
					Name:  `ГОСУДАРСТВЕННОЕ ОБЛАСТНОЕ АВТОНОМНОЕ УЧРЕЖДЕНИЕ СОЦИАЛЬНОГО ОБСЛУЖИВАНИЯ НАСЕЛЕНИЯ "СЕВЕРОМОРСКИЙ СПЕЦИАЛЬНЫЙ ДОМ ДЛЯ ОДИНОКИХ ПРЕСТАРЕЛЫХ"`,
					INN:   "5110001373",
					OKOPF: "20901 Автономные учреждения",
				},
				PurchaseMethod: "",
				Subject:        "поставка электрической энергии (мощности)",
				Services: []*agreement.AgreementService{
					{
						Name:            "поставка электрической энергии (мощности)",
						TypeObject:      "",
						Quantity:        200000,
						QuantityType:    "Киловатт-час",
						CountryOfOrigin: "Российская Федерация",
						OKPD:            "ОКДП: 4010419 Электроэнергия, произведенная электростанциями общего пользования прочими",
					},
				},
			},
		},
		{
			name:      "error",
			inputHTML: reader.ReadHtmlFile(dir + "/error.html"),
			inputData: &agreement.AgreementParesedData{
				ID: "", Number: "55110001373150000040000",
				Status: "Исполнение прекращено", Pfid: "3067",
				NoticeId: "", Price: 708602.24,
				SignedAt: getDate("10.06.2013"), ExecutionStart: getDate("10.06.2013"),
				ExecutionEnd: getDate("31.01.2016"), PublishedAt: getDate("16.01.2015"),
				UpdatedAt: getDate("02.03.2015"), Customer: &customer.Customer{ID: "87730"},
			},
			expectErr: false,
			expectedData: &agreement.AgreementParesedData{
				ID: "", Number: "55110001373150000040000",
				Status: "Исполнение прекращено", Pfid: "3067",
				NoticeId: "", Price: 708602.24,
				SignedAt: getDate("10.06.2013"), ExecutionStart: getDate("10.06.2013"),
				ExecutionEnd: getDate("31.01.2016"), PublishedAt: getDate("16.01.2015"),
				UpdatedAt: getDate("02.03.2015"), Customer: &customer.Customer{ID: "87730"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := agreement.ParseAgreementFromHtml(ctx, logger, tt.inputHTML, tt.inputData)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, result)
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
		inputData    *agreement.AgreementParesedData
		expectErr    bool
		expectedData *agreement.AgreementParesedData
	}{
		{
			name:      "correctTwoMainWork",
			inputHTML: reader.ReadHtmlFile(dir + "/correctTwoMainWork.html"),
			inputData: &agreement.AgreementParesedData{
				Customer: &customer.Customer{},
			},
			expectErr: false,
			expectedData: &agreement.AgreementParesedData{
				Customer: &customer.Customer{
					Location: "184601, МУРМАНСКАЯ, СЕВЕРОМОРСК, ГВАРДЕЙСКАЯ, дом 5",
					MainWork: "85.31: Предоставление социальных услуг с обеспечением проживания,87.90: Деятельность по уходу с обеспечением проживания прочая",
				},
			},
		},
		{
			name:      "corrcorrectOneMainWorkect",
			inputHTML: reader.ReadHtmlFile(dir + "/corrcorrectOneMainWorkect.html"),
			inputData: &agreement.AgreementParesedData{
				Customer: &customer.Customer{},
			},
			expectErr: false,
			expectedData: &agreement.AgreementParesedData{
				Customer: &customer.Customer{
					Location: "660022, Г. КРАСНОЯРСК, ОСТ-В ТАТЫШЕВ, ЗД. 2",
					MainWork: "96.04: Деятельность физкультурно- оздоровительная",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := agreement.ParseCustomerFromMain(ctx, logger, tt.inputHTML, tt.inputData)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedData, tt.inputData)
			}
		})
	}
}
