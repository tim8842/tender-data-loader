package model

// Variable - модель для хранения произвольных переменных.
type Variable struct {
	ID   string                 `bson:"_id,omitempty" json:"id"` // Уникальный идентификатор набора переменных.
	Vars map[string]interface{} `bson:"vars" json:"vars"`        // Произвольные переменные.
}
