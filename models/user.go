package models

// type User struct {
// 	ID       string `bson:"_id,omitempty" json:"id"`
// 	Name     string `bson:"name" json:"name"`
// 	Email    string `bson:"email" json:"email"`
// 	Password string `bson:"password" json:"password,omitempty"` // Allow in requests, omit in responses
// 	Role     string `bson:"role" json:"role"`
// 	SchoolID string `bson:"schoolId,omitempty" json:"schoolId,omitempty"` // keep if always expected
// }

type User struct {
	ID                 string           `bson:"_id" json:"_id"`
	Name               string           `bson:"name" json:"name"`
	Email              string           `bson:"email" json:"email"`
	Password           string           `bson:"password" json:"password"` // ✅ FIXED HERE
	Role               string           `bson:"role" json:"role"`
	SchoolID           string           `bson:"school_id,omitempty" json:"school_id,omitempty"`
	ClassroomIDs       []string         `bson:"classroomIds,omitempty" json:"classroomIds,omitempty"`
	SubscribedStudents []string         `bson:"subscribedStudents,omitempty" json:"subscribedStudents,omitempty"`
	SubscriptionMap    map[string]int64 `bson:"subscriptionMap,omitempty" json:"subscriptionMap,omitempty"`
}
