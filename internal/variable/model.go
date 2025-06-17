package variable

import (
	"encoding/json"
	"fmt"
	"time"
)

// Variable - модель для хранения произвольных переменных.
type Variable struct {
	ID   string                 `bson:"_id,omitempty" json:"id"` // Уникальный идентификатор набора переменных.
	Vars map[string]interface{} `bson:"vars" json:"vars"`        // Произвольные переменные.
}

func (t Variable) GetID() any {
	return t.ID
}

// Variable - модель для хранения произвольных переменных.
type VarsBackToNowAgreement struct {
	Page     int       `bson:"page,omitempty" json:"page"`
	SignedAt time.Time `bson:"signed_at,omitempty" json:"signed_at"`
}

type VariableBackToNowAgreement struct {
	ID   string                 `bson:"_id,omitempty" json:"id"` // Уникальный идентификатор набора переменных.
	Vars VarsBackToNowAgreement `bson:"vars" json:"vars"`        // Произвольные переменные.
}

func (t VariableBackToNowAgreement) ConvertToVariable() (Variable, error) {

	jsonData, err := json.Marshal(t)
	if err != nil {
		return Variable{}, fmt.Errorf("failed to marshal source struct: %w", err)
	}
	var temp Variable
	err = json.Unmarshal(jsonData, &temp)
	if err != nil {
		return Variable{}, fmt.Errorf("failed to unmarshal JSON into temp struct: %w", err)
	}
	return temp, nil
}
