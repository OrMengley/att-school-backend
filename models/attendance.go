package models

type Attendance struct {
	ID          string `bson:"_id" json:"id"`
	StudentID   string `bson:"studentId" json:"studentId"`
	ClassroomID string `bson:"classroomId" json:"classroomId"`
	Status      string `bson:"status" json:"status"` // "checkin" or "checkout"
	Timestamp   string `bson:"timestamp" json:"timestamp"`
}
