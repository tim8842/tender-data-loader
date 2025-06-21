package contract

import (
	"time"

	"github.com/tim8842/tender-data-loader/internal/customer"
	"github.com/tim8842/tender-data-loader/internal/supplier"
)

type ContractService struct {
	Name         string  `bson:"name" json:"name"` // наименование
	TypeObject   string  `bson:"type_object" json:"type_object"`
	Quantity     float64 `bson:"quantity" json:"quantity"` // количество
	QuantityType string  `bson:"quantity_type" json:"quantity_type"`
	UnitPrice    float64 `bson:"unit_price" json:"unit_price"` // цена за единицу
	Currency     string  `bson:"currency" json:"currency"`

	CountryOfOrigin   string `bson:"country_of_origin,omitempty" json:"country_of_origin,omitempty"`   // страна происхождения (опционально)
	CountryRegistered string `bson:"country_registered,omitempty" json:"country_registered,omitempty"` // страна регистрации (опционально)

	OKPD  string `bson:"okpd,omitempty" json:"okpd,omitempty"`   // ОКПД (опционально)
	OKPD2 string `bson:"okpd2,omitempty" json:"okpd2,omitempty"` // ОКПД2 (опционально)
}

type ContractParesedData struct {
	ID             string    `bson:"_id,omitempty" json:"id"` // номер договора (идентификатор)
	Status         string    `bson:"status,omitempty" json:"status"`
	NoticeId       string    `bson:"notice_id,omitempty" json:"notice_id"`
	Price          float64   `bson:"price" json:"price"`                     // цена договора
	SignedAt       time.Time `bson:"signed_at" json:"signed_at"`             // дата заключения
	ExecutionStart time.Time `bson:"execution_start" json:"execution_start"` // начало срока исполнения
	ExecutionEnd   time.Time `bson:"execution_end" json:"execution_end"`     // конец срока исполнения

	PublishedAt time.Time `bson:"published_at" json:"published_at"` // дата размещения
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`     // дата обновления
	Law         string    `bson:"law" json:"law"`                   // номер закона

	SupplierMethod string               `bson:"supplier_method" json:"supplier_method"` // способ закупки
	Subject        string               `bson:"subject" json:"subject"`                 // предмет договора
	Suppliers      []*supplier.Supplier `bson:"suppliers" json:"suppliers"`

	Customer *customer.Customer `bson:"customer" json:"customer"` // вложенный заказчик
	Services []*ContractService `bson:"services" json:"services"` // список услуг
}

type Contract struct {
	ID             string    `bson:"_id,omitempty" json:"id"` // номер договора (идентификатор)
	Status         string    `bson:"status,omitempty" json:"status"`
	NoticeId       string    `bson:"notice_id,omitempty" json:"notice_id"`
	Price          float64   `bson:"price" json:"price"`                     // цена договора
	SignedAt       time.Time `bson:"signed_at" json:"signed_at"`             // дата заключения
	ExecutionStart time.Time `bson:"execution_start" json:"execution_start"` // начало срока исполнения
	ExecutionEnd   time.Time `bson:"execution_end" json:"execution_end"`     // конец срока исполнения

	PublishedAt time.Time `bson:"published_at" json:"published_at"` // дата размещения
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`     // дата обновления
	Law         string    `bson:"law" json:"law"`                   // номер закона

	SupplierMethod string   `bson:"supplier_method" json:"supplier_method"` // способ закупки
	Subject        string   `bson:"subject" json:"subject"`                 // предмет договора
	SupplierIds    []string `bson:"supplier_ids" json:"supplier_ids"`

	CustomerId string             `bson:"customer_id" json:"customer_id"` // вложенный заказчик
	Services   []*ContractService `bson:"services" json:"services"`       // список услуг
}

func (t Contract) GetID() any {
	return t.ID
}

func ParseContractParsedDataToModel(data *ContractParesedData) (*Contract, *customer.Customer, []*supplier.Supplier) {
	supplierIDs := make([]string, 0, len(data.Suppliers))
	suppliers := make([]*supplier.Supplier, 0, len(data.Suppliers))

	for _, s := range data.Suppliers {
		if s != nil && s.ID != "" {
			supplierIDs = append(supplierIDs, s.ID)
			suppliers = append(suppliers, s) // добавляем весь объект
		}
	}

	contract := &Contract{
		ID:             data.ID,
		Status:         data.Status,
		NoticeId:       data.NoticeId,
		Price:          data.Price,
		SignedAt:       data.SignedAt,
		ExecutionStart: data.ExecutionStart,
		ExecutionEnd:   data.ExecutionEnd,
		PublishedAt:    data.PublishedAt,
		UpdatedAt:      data.UpdatedAt,
		Law:            data.Law,
		SupplierMethod: data.SupplierMethod,
		Subject:        data.Subject,
		SupplierIds:    supplierIDs,
		CustomerId:     data.Customer.ID,
		Services:       data.Services,
	}

	customer := &customer.Customer{
		ID:       data.Customer.ID,
		Code:     data.Customer.Code,
		Name:     data.Customer.Name,
		INN:      data.Customer.INN,
		OKOPF:    data.Customer.OKOPF,
		MainWork: data.Customer.MainWork,
		Location: data.Customer.Location,
	}

	return contract, customer, suppliers
}
