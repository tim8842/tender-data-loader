package model

type Customer struct {
	ID           string `bson:"_id,omitempty" json:"id"`            // Mongo ID
	Name         string `bson:"name" json:"name"`                   // Название заказчика
	URL          string `bson:"url" json:"url"`                     // URL заказчика
	INN          string `bson:"inn" json:"inn"`                     // ИНН
	CustomerCode string `bson:"customer_code" json:"customer_code"` // Идентификационный код заказчика
	OKOPF        string `bson:"okopf" json:"okopf"`                 // Организационно-правовая форма
	Location     string `bson:"location" json:"location"`           // Место нахождения
}
