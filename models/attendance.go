package models

import "time"

type Attendance struct {
	ID          string    `bson:"_id" json:"_id"`
	StudentID   string    `bson:"studentId" json:"studentId"`
	ClassroomID string    `bson:"classroomId" json:"classroomId"`
	Date        string    `bson:"date" json:"date"`     // YYYY-MM-DD
	Status      string    `bson:"status" json:"status"` // "checkin" or "checkout"
	Time        time.Time `bson:"time" json:"time"`
	TeacherID   string    `bson:"teacherId" json:"teacherId"`
}
