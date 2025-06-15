package models

type School struct {
	ID      string `bson:"_id" json:"id"`
	Name    string `bson:"name" json:"name"`
	Address string `bson:"address" json:"address"`
}
