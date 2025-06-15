package models

type Department struct {
	ID       string `bson:"_id" json:"id"`
	Name     string `bson:"name" json:"name"`
	SchoolID string `bson:"schoolId" json:"schoolId"`
}
