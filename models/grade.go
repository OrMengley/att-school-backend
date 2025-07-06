package models

import "time"

type Grade struct {
	ID        string    `bson:"_id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	TermID    string    `bson:"termId" json:"termId"`
	SchoolID  string    `bson:"schoolId,omitempty" json:"schoolId,omitempty"` // ✅ NEW
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
