package models

type Department struct {
	ID       string `bson:"_id" json:"_id"`
	Name     string `bson:"name" json:"name"`
	SchoolID string `bson:"school_id" json:"school_id"`
}
