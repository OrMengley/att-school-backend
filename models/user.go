package models

type User struct {
	ID       string `bson:"_id,omitempty" json:"id"`
	Name     string `bson:"name" json:"name"`
	Email    string `bson:"email" json:"email"`
	Password string `bson:"password" json:"password,omitempty"` // Allow in requests, omit in responses
	Role     string `bson:"role" json:"role"`
	SchoolID string `bson:"schoolId,omitempty" json:"schoolId,omitempty"` // keep if always expected
}
