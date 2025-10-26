package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"school-attendance/db"
	"school-attendance/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // ✅ new import
	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateClassroom handles the creation of a new classroom
// It accepts a JSON payload with name and gradeId
// Returns the newly created classroom ID on success
func CreateClassroom(c *gin.Context) {

	type Input struct {
		ID           string `json:"_id"`
		Name         string `json:"name"`
		GradeID      string `json:"grade_id"`
		TermID       string `json:"term_id"`
		SchoolID     string `json:"school_id"`
		Status       string `json:"status"`
		DepartmentID string `json:"department_id"`
	}

	var input Input
	// if err := c.ShouldBindJSON(&input); err != nil || input.ID == "" || input.GradeID == "" {
	if err := c.ShouldBindJSON(&input); err != nil || input.GradeID == "" || input.DepartmentID == "" || input.SchoolID == "" || input.Name == "" || input.Status == "" || input.TermID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom input"})
		return
	}
	id := "classroom-" + uuid.New().String() // ✅ Auto UUID
	classroom := bson.M{
		"_id":           id,
		"name":          input.Name,
		"grade_id":      input.GradeID,
		"status":        input.Status,
		"department_id": input.DepartmentID,
		"school_id":     input.SchoolID,
		"term_id":       input.TermID,
		"createdAt":     time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")
	_, err := col.InsertOne(ctx, classroom)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create classroom"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Classroom created", "id": id})
}

// GetClassrooms retrieves all classrooms from the database
// Returns a list of all classrooms as JSON
func GetClassrooms(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	cursor, err := col.Aggregate(ctx, bson.A{

		bson.M{
			"$lookup": bson.M{
				"from":         "grades",
				"localField":   "grade_id",
				"foreignField": "_id",
				"as":           "grade",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "terms",
				"localField":   "term_id",
				"foreignField": "_id",
				"as":           "term",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$term",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$grade",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})
	fmt.Println(cursor)
	defer cancel()

	// cursor, err := col.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classrooms"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode classrooms"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetClassroom retrieves a single classroom by ID from the database
// Returns the classroom as JSON or 404 if not found
func GetClassroom(c *gin.Context) {
	// Get classroom ID from URL parameter
	classroomID := c.Param("id")
	if classroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Classroom ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	// Use aggregation to get classroom with grade and term information
	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": classroomID,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "grades",
				"localField":   "grade_id",
				"foreignField": "_id",
				"as":           "grade",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "terms",
				"localField":   "term_id",
				"foreignField": "_id",
				"as":           "term",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$term",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$grade",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classroom"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode classroom"})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	c.JSON(http.StatusOK, results[0])
}

// UpdateClassroom handles updating an existing classroom
// It accepts a JSON payload with classroom ID and updated fields
// Returns success message on successful update

// UpdateClassroom handles updating an existing classroom
func UpdateClassroom(c *gin.Context) {
	classroomID := c.Param("id")
	if classroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Classroom ID is required"})
		return
	}
	// Log the request for debugging
	fmt.Printf("Updating classroom with ID: %s\n", classroomID)

	var input models.Classroom
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom input"})
		fmt.Printf("Invalid input error: %v\n", err)
		return
	}

	// Build update fields
	updateFields := bson.M{}

	if input.Name != "" {
		updateFields["name"] = input.Name
	}
	if input.GradeID != "" {
		updateFields["grade_id"] = input.GradeID
	}
	if input.TermID != "" {
		updateFields["term_id"] = input.TermID
	}
	// if input.StudentIDs != nil {
	// 	updateFields["studentIds"] = input.StudentIDs
	// }
	if input.Sessions != nil {
		var sessions []models.Session
		for _, session := range input.Sessions {
			sessions = append(sessions, models.Session{
				ID:               uuid.New().String(),
				SubjectName:      session.SubjectName,
				TeacherID:        session.TeacherID,
				SessionDay:       session.SessionDay,
				SessionStartHour: session.SessionStartHour,
				SessionStartMin:  session.SessionStartMin,
				SessionEndHour:   session.SessionEndHour,
				SessionEndMin:    session.SessionEndMin,
			})
		}
		updateFields["sessions"] = sessions
	}

	// Only update if there are fields to update
	if len(updateFields) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	fmt.Printf("Update fields: %+v\n", updateFields)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	// First check if the classroom exists
	var existingClassroom bson.M
	err := col.FindOne(ctx, bson.M{"_id": classroomID}).Decode(&existingClassroom)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		fmt.Printf("Classroom not found: %s, error: %v\n", classroomID, err)
		return
	}

	// Proceed with update
	result, err := col.UpdateOne(
		ctx,
		bson.M{"_id": classroomID},
		bson.M{"$set": updateFields},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update classroom"})
		fmt.Printf("Update error: %v\n", err)
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Classroom updated successfully"})

}

func AssignStudentsToClassroom(c *gin.Context) {
	type Input struct {
		StudentIDs  []string `json:"studentIds"`
		TermID      string   `json:"term_id"`
		ClassroomID string   `json:"classroom_id"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil || input.TermID == "" || input.ClassroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("students")

	for _, studentID := range input.StudentIDs {
		filter := bson.M{"_id": studentID}

		update := bson.M{
			"$set": bson.M{
				"term_id":     input.TermID,
				"classroomId": input.ClassroomID,
			},
			"$addToSet": bson.M{
				"enrollments": bson.M{
					"term_id":      input.TermID,
					"classroom_id": input.ClassroomID,
				},
			},
		}

		_, err := col.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update student %s", studentID)})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Students assigned successfully"})
}

// AssignTeacherToClassroom assigns a teacher to a classroom by updating the teacher's classroomIds array
func AssignTeacherToClassroom(c *gin.Context) {
	type Input struct {
		TeacherID   string `json:"teacher_id"`
		ClassroomID string `json:"classroom_id"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil || input.TeacherID == "" || input.ClassroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("users")

	filter := bson.M{"_id": input.TeacherID, "role": "teacher"}
	update := bson.M{
		"$addToSet": bson.M{
			"classroomIds": input.ClassroomID,
		},
	}

	result, err := col.UpdateOne(ctx, filter, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign teacher to classroom"})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Teacher not found or not a teacher role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Teacher assigned to classroom successfully"})
}

// AddSessionToClassroom adds a single session to an existing classroom
// AddSessionToClassroom adds one or more sessions to an existing classroom
func AddSessionToClassroom(c *gin.Context) {
	classroomID := c.Param("id")
	if classroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Classroom ID is required"})
		return
	}

	// Input: array of sessions
	var input []struct {
		SubjectName      string `json:"subjectName"`
		TeacherID        string `json:"teacher_id"`
		SessionDay       string `json:"sessionDay"`
		SessionStartHour string `json:"sessionStartHour"`
		SessionStartMin  string `json:"sessionStartMin"`
		SessionEndHour   string `json:"sessionEndHour"`
		SessionEndMin    string `json:"sessionEndMin"`
		SessionStatus    string `json:"sessionStatus"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session input"})
		return
	}

	if len(input) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No sessions provided"})
		return
	}

	// Convert to []models.Session with generated UUID
	var newSessions []models.Session
	for _, s := range input {
		newSessions = append(newSessions, models.Session{
			ID:               uuid.New().String(),
			SubjectName:      s.SubjectName,
			TeacherID:        s.TeacherID,
			SessionDay:       s.SessionDay,
			SessionStartHour: s.SessionStartHour,
			SessionStartMin:  s.SessionStartMin,
			SessionEndHour:   s.SessionEndHour,
			SessionEndMin:    s.SessionEndMin,
			SessionStatus:    s.SessionStatus,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")
	// 1. Check for required fields and time correctness
	for _, s := range newSessions {
		if s.SubjectName == "" || s.TeacherID == "" || s.SessionDay == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
			return
		}

		startH, _ := strconv.Atoi(s.SessionStartHour)
		startM, _ := strconv.Atoi(s.SessionStartMin)
		endH, _ := strconv.Atoi(s.SessionEndHour)
		endM, _ := strconv.Atoi(s.SessionEndMin)

		start := startH*60 + startM
		end := endH*60 + endM
		if start >= end {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Start time must be before end time"})
			return
		}
	}

	// 2. Fetch existing sessions for this classroom
	var existing struct {
		Sessions []models.Session `bson:"sessions"`
	}
	err := col.FindOne(ctx, bson.M{"_id": classroomID}).Decode(&existing)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	// 3. Validate against overlapping sessions for the same day
	mustParse := func(hh, mm string) int {
		h, _ := strconv.Atoi(hh)
		m, _ := strconv.Atoi(mm)
		return h*60 + m
	}

	for _, ns := range newSessions {
		nsStart := mustParse(ns.SessionStartHour, ns.SessionStartMin)
		nsEnd := mustParse(ns.SessionEndHour, ns.SessionEndMin)

		for _, es := range existing.Sessions {
			if es.SessionDay == ns.SessionDay {
				esStart := mustParse(es.SessionStartHour, es.SessionStartMin)
				esEnd := mustParse(es.SessionEndHour, es.SessionEndMin)
				// Check overlap
				if nsStart < esEnd && nsEnd > esStart {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": fmt.Sprintf("Session conflict on %s with %s", ns.SessionDay, es.SubjectName),
					})
					return
				}
			}
		}
	}

	// ------------------- END VALIDATION -------------------

	// Use $push with $each to append multiple sessions
	update := bson.M{
		"$push": bson.M{
			"sessions": bson.M{
				"$each": newSessions,
			},
		},
	}

	result, err := col.UpdateOne(ctx, bson.M{"_id": classroomID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add sessions"})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Sessions added successfully",
		"sessions": newSessions,
	})
}

// GetSessionsForClassroom retrieves only the sessions for a given classroom
func GetSessionsForClassroom(c *gin.Context) {
	classroomID := c.Param("id")
	if classroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Classroom ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	// Projection to only include sessions field
	var result struct {
		Sessions []models.Session `bson:"sessions" json:"sessions"`
	}

	err := col.FindOne(
		ctx,
		bson.M{"_id": classroomID},
		options.FindOne().SetProjection(bson.M{"sessions": 1, "_id": 0}),
	).Decode(&result)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	c.JSON(http.StatusOK, result.Sessions)
}

// RemoveSessionFromClassroom removes a session by its sessionId from a classroom
func RemoveSessionFromClassroom(c *gin.Context) {
	classroomID := c.Param("id")
	sessionID := c.Param("sessionId")

	if classroomID == "" || sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Classroom ID and Session ID are required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	// Use $pull to remove the session with this ID
	update := bson.M{
		"$pull": bson.M{
			"sessions": bson.M{"_id": sessionID},
		},
	}

	result, err := col.UpdateOne(ctx, bson.M{"_id": classroomID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove session"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Session removed successfully"})
}

// GetClassroomsBySchoolID retrieves all classrooms for a specific school

// Returns a list of classrooms filtered by school_id as JSON
func GetClassroomsBySchoolID(c *gin.Context) {
	schoolID := c.Param("school_id")
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"school_id": schoolID,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "grades",
				"localField":   "grade_id",
				"foreignField": "_id",
				"as":           "grade",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "terms",
				"localField":   "term_id",
				"foreignField": "_id",
				"as":           "term",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$term",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$grade",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classrooms"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode classrooms"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetClassroomByTermID retrieves all classrooms for a specific term
// Returns a list of classrooms filtered by term_id as JSON
func GetClassroomByTermID(c *gin.Context) {
	termID := c.Param("term_id")
	if termID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"term_id": termID,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "grades",
				"localField":   "grade_id",
				"foreignField": "_id",
				"as":           "grade",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "terms",
				"localField":   "term_id",
				"foreignField": "_id",
				"as":           "term",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$term",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$grade",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch classrooms"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode classrooms"})
		return
	}

	c.JSON(http.StatusOK, results)
}
