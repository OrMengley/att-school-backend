package models

type Session struct {
	ID               string `bson:"_id" json:"_id"`
	SubjectName      string `bson:"subjectName" json:"subjectName"`
	TeacherID        string `bson:"teacher_id" json:"teacher_id"`
	SessionDay       string `bson:"sessionDay" json:"sessionDay"`
	SessionStartHour string `bson:"sessionStartHour" json:"sessionStartHour"`
	SessionStartMin  string `bson:"sessionStartMin" json:"sessionStartMin"`
	SessionEndHour   string `bson:"sessionEndHour" json:"sessionEndHour"`
	SessionEndMin    string `bson:"sessionEndMin" json:"sessionEndMin"`
	SessionStatus    string `bson:"sessionStatus,omitempty" json:"sessionStatus,omitempty"`
}

type Classroom struct {
	ID           string    `bson:"_id" json:"_id"`
	Name         string    `bson:"name" json:"name"`
	GradeID      string    `bson:"grade_id" json:"grade_id"`
	TermID       string    `bson:"term_id" json:"term_id"`
	DepartmentID string    `bson:"department_id" json:"department_id"`
	SchoolID     string    `bson:"school_id" json:"school_id"`
	CreatedAt    string    `bson:"createdAt" json:"createdAt"`
	Status       string    `bson:"status,omitempty" json:"status,omitempty"`
	Sessions     []Session `bson:"sessions,omitempty" json:"sessions,omitempty"`
}
