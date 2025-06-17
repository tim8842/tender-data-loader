package repository_test

type TestEntity struct {
	ID    string `bson:"_id" json:"id"`
	Value string `bson:"value" json:"value"`
}

func (t TestEntity) GetID() any {
	return t.ID
}
