package model

type AgreementService struct {
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
