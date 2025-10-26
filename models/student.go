package models

type Enrollment struct {
	TermID      string `bson:"term_id" json:"term_id"`
	ClassroomID string `bson:"classroom_id" json:"classroom_id"`
}

type Student struct {
	ID          string       `bson:"_id" json:"_id"`
	SchoolID    string       `bson:"school_id" json:"school_id"`
	FirstName   string       `bson:"firstName,omitempty" json:"firstName,omitempty"` // optional
	LastName    string       `bson:"lastName,omitempty" json:"lastName,omitempty"`   // optional
	Name        string       `bson:"name" json:"name"`
	Sex         string       `json:"sex"`                                              // required
	DateOfBirth string       `bson:"dob,omitempty" json:"dob,omitempty"`               // optional, format: YYYY-MM-DD
	Address     string       `bson:"address,omitempty" json:"address,omitempty"`       // optional
	ProfileURL  string       `bson:"profileUrl,omitempty" json:"profileUrl,omitempty"` // optional
	ParentPhone string       `bson:"parentPhone" json:"parentPhone"`                   // required
	Enrollments []Enrollment `bson:"enrollments" json:"enrollments"`                   // each term & class
	Status      string       `bson:"status" json:"status"`
	CreatedAt   string       `bson:"createdAt" json:"createdAt"`
	ClassroomID string       `bson:"classroom_id,omitempty" json:"classroom_id,omitempty"`
	TermID      string       `bson:"term_id,omitempty" json:"term_id,omitempty"`
	IdNumber    string       `bson:"id_number,omitempty" json:"id_number,omitempty"` // optional
}
