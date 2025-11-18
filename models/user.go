package models

type User struct {
	ID                 string           `bson:"_id" json:"_id"`
	Name               string           `bson:"name" json:"name"`
	Email              string           `bson:"email" json:"email"`
	Password           string           `bson:"password" json:"password"` // ✅ FIXED HERE
	Role               string           `bson:"role" json:"role"`
	Status             string           `bson:"status,omitempty" json:"status,omitempty"`
	SchoolID           string           `bson:"school_id,omitempty" json:"school_id,omitempty"`
	Enrollments        []Enrollment     `bson:"enrollments, omitempty" json:"enrollments"`
	ClassroomIDs       []string         `bson:"classroomIds,omitempty" json:"classroomIds,omitempty"`
	SubscribedStudents []string         `bson:"subscribedStudents,omitempty" json:"subscribedStudents,omitempty"`
	SubscriptionMap    map[string]int64 `bson:"subscriptionMap,omitempty" json:"subscriptionMap,omitempty"`
}
