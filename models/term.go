package models

type Term struct {
	ID           string `bson:"_id" json:"id"`
	Year         string `bson:"year" json:"year"`
	DepartmentID string `bson:"departmentId" json:"departmentId"`
}
