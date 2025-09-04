package models

type Term struct {
	ID           string `bson:"_id" json:"_id"`
	Year         string `bson:"year" json:"year"`
	DepartmentID string `bson:"department_id" json:"department_id"`
}
