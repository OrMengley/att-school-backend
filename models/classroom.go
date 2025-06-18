package models

type Classroom struct {
	ID      string `bson:"_id" json:"id"`
	Name    string `bson:"name" json:"name"`
	GradeID string `bson:"gradeId" json:"gradeId"`
}