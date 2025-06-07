package model

type Customer struct {
	ID       string `bson:"_id,omitempty" json:"id"` // Mongo ID
	Code     string `bson:"code" json:"code"`
	Name     string `bson:"name" json:"name"`   // Название заказчика
	INN      string `bson:"inn" json:"inn"`     // ИНН
	OKOPF    string `bson:"okopf" json:"okopf"` // Организационно-правовая форма
	MainWork string `bson:"main_work" json:"main_work"`
	Location string `bson:"location" json:"location"` // Место нахождения
}
