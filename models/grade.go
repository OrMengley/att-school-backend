package models

type Grade struct {
	ID     string `bson:"_id" json:"id"`
	Name   string `bson:"name" json:"name"`
	TermID string `bson:"termId" json:"termId"`
}
