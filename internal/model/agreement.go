package model

import "time"

type AgreementParesedData struct {
	ID             string    `bson:"_id,omitempty" json:"id"` // номер договора (идентификатор)
	Number         string    `bson:"number,omitempty" json:"number"`
	Status         string    `bson:"status,omitempty" json:"status"`
	Pfid           string    `bson:"pfid,omitempty" json:"pdif"`
	NoticeId       string    `bson:"notice_id,omitempty" json:"notice_id"`
	Price          float64   `bson:"price" json:"price"`                     // цена договора
	SignedAt       time.Time `bson:"signed_at" json:"signed_at"`             // дата заключения
	ExecutionStart time.Time `bson:"execution_start" json:"execution_start"` // начало срока исполнения
	ExecutionEnd   time.Time `bson:"execution_end" json:"execution_end"`     // конец срока исполнения

	PublishedAt time.Time `bson:"published_at" json:"published_at"` // дата размещения
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`     // дата обновления

	PurchaseMethod string `bson:"purchase_method" json:"purchase_method"` // способ закупки
	Subject        string `bson:"subject" json:"subject"`                 // предмет договора

	Customer *Customer           `bson:"customer" json:"customer"` // вложенный заказчик
	Services []*AgreementService `bson:"services" json:"services"` // список услуг
}

type Agreement struct {
	ID             string    `bson:"_id,omitempty" json:"id"` // номер договора (идентификатор)
	Number         string    `bson:"number,omitempty" json:"number"`
	Status         string    `bson:"status,omitempty" json:"status"`
	Pfid           string    `bson:"pfid,omitempty" json:"pdif"`
	NoticeId       string    `bson:"notice_id,omitempty" json:"notice_id"`
	Price          float64   `bson:"price" json:"price"`                     // цена договора
	SignedAt       time.Time `bson:"signed_at" json:"signed_at"`             // дата заключения
	ExecutionStart time.Time `bson:"execution_start" json:"execution_start"` // начало срока исполнения
	ExecutionEnd   time.Time `bson:"execution_end" json:"execution_end"`     // конец срока исполнения

	PublishedAt time.Time `bson:"published_at" json:"published_at"` // дата размещения
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`     // дата обновления

	PurchaseMethod string `bson:"purchase_method" json:"purchase_method"` // способ закупки
	Subject        string `bson:"subject" json:"subject"`                 // предмет договора

	CustomerId string              `bson:"customer_id" json:"customer_id"` // вложенный заказчик
	Services   []*AgreementService `bson:"services" json:"services"`       // список услуг
}

func (t Agreement) GetID() any {
	return t.ID
}

func ParseAgreementDataToModels(data *AgreementParesedData) (*Agreement, *Customer) {
	agreement := &Agreement{
		ID:             data.ID,
		Number:         data.Number,
		Status:         data.Status,
		Pfid:           data.Pfid,
		NoticeId:       data.NoticeId,
		Price:          data.Price,
		SignedAt:       data.SignedAt,
		ExecutionStart: data.ExecutionStart,
		ExecutionEnd:   data.ExecutionEnd,
		PublishedAt:    data.PublishedAt,
		UpdatedAt:      data.UpdatedAt,
		PurchaseMethod: data.PurchaseMethod,
		Subject:        data.Subject,
		CustomerId:     data.Customer.ID, // связываем ID
		Services:       data.Services,
	}

	customer := &Customer{
		ID:       data.Customer.ID,
		Code:     data.Customer.Code,
		Name:     data.Customer.Name,
		INN:      data.Customer.INN,
		OKOPF:    data.Customer.OKOPF,
		MainWork: data.Customer.MainWork,
		Location: data.Customer.Location,
	}

	return agreement, customer
}
