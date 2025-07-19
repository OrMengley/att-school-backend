package models

type Session struct {
	ID               string `bson:"_id" json:"id"`
	SubjectName      string `bson:"subjectName" json:"subjectName"`
	TeacherID        string `bson:"teacherId" json:"teacherId"`
	SessionDay       string `bson:"sessionDay" json:"sessionDay"`
	SessionStartHour string `bson:"sessionStartHour" json:"sessionStartHour"`
	SessionStartMin  string `bson:"sessionStartMin" json:"sessionStartMin"`
	SessionEndHour   string `bson:"sessionEndHour" json:"sessionEndHour"`
	SessionEndMin    string `bson:"sessionEndMin" json:"sessionEndMin"`
}

type Classroom struct {
	ID        string    `bson:"_id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	GradeID   string    `bson:"gradeId" json:"gradeId"`
	TermID    string    `bson:"termId" json:"termId"`
	CreatedAt string    `bson:"createdAt" json:"createdAt"`
	Sessions  []Session `bson:"sessions,omitempty" json:"sessions,omitempty"`
}
