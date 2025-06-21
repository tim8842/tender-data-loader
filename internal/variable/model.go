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

// VariableBackToNow .
type VarsBackToNow struct {
	Page     int       `bson:"page,omitempty" json:"page"`
	SignedAt time.Time `bson:"signed_at,omitempty" json:"signed_at"`
}

type VariableBackToNow struct {
	ID   string        `bson:"_id,omitempty" json:"id"` // Уникальный идентификатор набора переменных.
	Vars VarsBackToNow `bson:"vars" json:"vars"`        // Произвольные переменные.
}

func (t VariableBackToNow) ConvertToVariable() (Variable, error) {
	return convertToVariable(t)
}

func convertToVariable(t any) (Variable, error) {
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

// VarsBackToNowContract .
type VarsBackToNowContract struct {
	Page      int       `bson:"page,omitempty" json:"page"`
	Fz        string    `bson:"fz,omitempty" json:"fz"`
	SignedAt  time.Time `bson:"signed_at,omitempty" json:"signed_at"`
	PriceFrom float32   `bson:"price_from,omitempty" json:"price_from"`
	PriceTo   float32   `bson:"price_to,omitempty" json:"price_to"`
}

type VariableBackToNowContract struct {
	ID   string                `bson:"_id,omitempty" json:"id"`
	Vars VarsBackToNowContract `bson:"vars" json:"vars"`
}

func (t VariableBackToNowContract) ConvertToVariable() (Variable, error) {
	return convertToVariable(t)
}
