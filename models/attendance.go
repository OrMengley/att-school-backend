package models

import "time"

type Attendance struct {
	ID           string    `bson:"_id" json:"id"`
	StudentID    string    `bson:"studentId" json:"studentId"`
	ClassroomID  string    `bson:"classroomId" json:"classroomId"`
	TeacherID    string    `bson:"teacherId" json:"teacherId"`
	SessionID    string    `bson:"sessionId" json:"sessionId"` // ✅ New field
	Date         string    `bson:"date" json:"date"`           // YYYY-MM-DD
	Status       string    `bson:"status" json:"status"`       // "checkin", "checkout"
	CheckinTime  time.Time `bson:"checkinTime,omitempty" json:"checkinTime,omitempty"`
	CheckoutTime time.Time `bson:"checkoutTime,omitempty" json:"checkoutTime,omitempty"`
}
