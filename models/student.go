package models

type Student struct {
	ID          string `bson:"_id" json:"id"`
	Name        string `bson:"name" json:"name"`
	ClassroomID string `bson:"classroomId" json:"classroomId"`
	ParentPhone string `bson:"parentPhone" json:"parentPhone"`
}
