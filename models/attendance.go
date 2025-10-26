package models

import "time"

type Attendance struct {
	ID           string    `bson:"_id" json:"_id"`
	StudentID    string    `bson:"student_id" json:"student_id"`
	ClassroomID  string    `bson:"classroom_id" json:"classroom_id"`
	TeacherID    string    `bson:"teacher_id" json:"teacher_id"`
	TermID       string    `bson:"term_id,omitempty" json:"term_id,omitempty"`
	DepartmentID string    `bson:"department_id,omitempty" json:"department_id,omitempty"`
	SchoolID     string    `bson:"school_id,omitempty" json:"school_id,omitempty"`
	SessionID    string    `bson:"session_id" json:"session_id"` // ✅ New field
	Date         string    `bson:"date" json:"date"`             // YYYY-MM-DD
	Status       string    `bson:"status" json:"status"`         // "checkin", "checkout"
	CheckinTime  time.Time `bson:"checkinTime,omitempty" json:"checkinTime,omitempty"`
	CheckoutTime time.Time `bson:"checkoutTime,omitempty" json:"checkoutTime,omitempty"`
}
