package models

import "time"

type Term struct {
	ID           string     `bson:"_id" json:"_id"`
	Year         string     `bson:"year" json:"year"`
	DepartmentID string     `bson:"department_id" json:"department_id"`
	Deleted      bool       `bson:"deleted,omitempty" json:"deleted"`
	DeletedAt    *time.Time `bson:"deletedAt,omitempty" json:"deletedAt,omitempty"`
	CreatedAt    *time.Time `bson:"createdAt,omitempty" json:"createdAt,omitempty"`
}
