package models

import "time"

type Grade struct {
	ID        string    `bson:"_id" json:"_id"`
	Name      string    `bson:"name" json:"name"`
	TermID    string    `bson:"term_id" json:"term_id"`
	SchoolID  string    `bson:"school_id,omitempty" json:"school_id,omitempty"` // ✅ NEW
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
