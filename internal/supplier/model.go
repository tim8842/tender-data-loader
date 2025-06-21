package supplier

type Supplier struct {
	ID           string `bson:"_id,omitempty" json:"id"` // Mongo ID
	Name         string `bson:"name" json:"name"`        // Название заказчика
	Country      string `bson:"country" json:"country"`
	INN          string `bson:"inn" json:"inn"`                     // ИНН
	Location     string `bson:"location" json:"location"`           // Место нахождения
	MailLocation string `bson:"mail_location" json:"mail_location"` // Место нахождения
	Contact      string `bson:"contact" json:"contact"`             // Место нахождения
	Status       string `bson:"status" json:"status"`               // Место нахождения
}

func (t Supplier) GetID() any {
	return t.ID
}
