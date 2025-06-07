package model

import "time"

// Variable - модель для хранения произвольных переменных.
type Variable struct {
	ID   string                 `bson:"_id,omitempty" json:"id"` // Уникальный идентификатор набора переменных.
	Vars map[string]interface{} `bson:"vars" json:"vars"`        // Произвольные переменные.
}

// Variable - модель для хранения произвольных переменных.
type VarsBackToNowAgreement struct {
	Page      int       `bson:"page,omitempty" json:"page"`
	Signed_at time.Time `bson:"signed_at,omitempty" json:"signed_at"`
}

type VariableBackToNowAgreement struct {
	ID   string                 `bson:"_id,omitempty" json:"id"` // Уникальный идентификатор набора переменных.
	Vars VarsBackToNowAgreement `bson:"vars" json:"vars"`        // Произвольные переменные.
}
