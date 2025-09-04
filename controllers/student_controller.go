package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"school-attendance/db"
)

func CreateStudent(c *gin.Context) {
	type Input struct {
		FirstName   string `json:"firstName,omitempty"`
		LastName    string `json:"lastName,omitempty"`
		Name        string `json:"name"` // Required
		DateOfBirth string `json:"dob,omitempty"`
		Address     string `json:"address,omitempty"`
		ProfileURL  string `json:"profileUrl,omitempty"`
		Sex         string `json:"sex,omitempty"`         // Optional: "male", "female", etc.
		ParentPhone string `json:"parentPhone"`           // Required
		ClassroomID string `json:"classroomId,omitempty"` // Optional
		TermID      string `json:"termId,omitempty"`      // Optional
		Status      string `json:"status,omitempty"`      // Optional: "active" or "inactive"
		Enrollments []struct {
			TermID      string `json:"termId,omitempty"`
			ClassroomID string `json:"classroomId,omitempty"`
		} `json:"enrollments,omitempty"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student input"})
		return
	}

	// Default status to "active" if not provided
	status := input.Status
	if status == "" {
		status = "active"
	}

	id := "student-" + uuid.New().String()
	createdAt := time.Now()

	// Build enrollment if term or classroom is provided
	var enrollments []bson.M
	if input.TermID != "" || input.ClassroomID != "" {
		enrollment := bson.M{}
		if input.TermID != "" {
			enrollment["termId"] = input.TermID
		}
		if input.ClassroomID != "" {
			enrollment["classroomId"] = input.ClassroomID
		}
		enrollments = append(enrollments, enrollment)
	}

	student := bson.M{
		"_id":         id,
		"firstName":   input.FirstName,
		"lastName":    input.LastName,
		"name":        input.Name,
		"dob":         input.DateOfBirth,
		"address":     input.Address,
		"profileUrl":  input.ProfileURL,
		"sex":         input.Sex,
		"parentPhone": input.ParentPhone,
		"status":      status,
		"classroomId": input.ClassroomID,
		"termId":      input.TermID,
		"enrollments": enrollments,
		"createdAt":   createdAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("students")
	_, err := col.InsertOne(ctx, student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Student created",
		"id":      id,
	})
}

// func UpdateStudentClassroom(c *gin.Context) {
// 	id := c.Param("id")

// 	type Input struct {
// 		ClassroomID string `json:"classroomId"`
// 		TermID      string `json:"termId"`
// 	}

// 	var input Input
// 	if err := c.ShouldBindJSON(&input); err != nil || input.ClassroomID == "" || input.TermID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom update input"})
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	col := db.MongoClient.Database("school_attendance").Collection("students")

// 	update := bson.M{
// 		"$set": bson.M{
// 			"classroomId": input.ClassroomID,
// 			"termId":      input.TermID,
// 			"updatedAt":   time.Now(),
// 		},
// 		"$push": bson.M{
// 			"enrollments": bson.M{
// 				"classroomId": input.ClassroomID,
// 				"termId":      input.TermID,
// 			},
// 		},
// 	}

// 	_, err := col.UpdateOne(ctx, bson.M{"_id": id}, update)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student classroom"})
// 		return
// 	}

//		c.JSON(http.StatusOK, gin.H{"message": "Student classroom updated"})
//	}
// func UpdateStudentClassroom(c *gin.Context) {
// 	id := c.Param("id")

// 	type Input struct {
// 		ClassroomID string `json:"classroomId"`
// 		TermID      string `json:"termId"`
// 	}

// 	var input Input
// 	if err := c.ShouldBindJSON(&input); err != nil || input.ClassroomID == "" || input.TermID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom update input"})
// 		return
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	col := db.MongoClient.Database("school_attendance").Collection("students")

// 	update := bson.M{
// 		"$set": bson.M{
// 			"classroomId": input.ClassroomID,
// 			"termId":      input.TermID,
// 			"updatedAt":   time.Now(),
// 		},
// 		"$addToSet": bson.M{
// 			"enrollments": bson.M{
// 				"classroomId": input.ClassroomID,
// 				"termId":      input.TermID,
// 			},
// 		},
// 	}

// 	_, err := col.UpdateOne(ctx, bson.M{"_id": id}, update)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student classroom"})
// 		return
// 	}

//		c.JSON(http.StatusOK, gin.H{"message": "Student classroom updated"})
//	}

// function to update student classroom and term and add new enrollment
func UpdateStudentClassroom(c *gin.Context) {
	id := c.Param("id")

	type Input struct {
		ClassroomID string `json:"classroomId"`
		TermID      string `json:"termId"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil || input.ClassroomID == "" || input.TermID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom update input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("students")

	// First check if student exists and if classroomId, termId, and enrollments are empty
	var student bson.M
	err := col.FindOne(ctx, bson.M{"_id": id}).Decode(&student)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// Check if classroomId and termId are empty and enrollments is null/empty
	classroomID, classroomExists := student["classroomId"]
	termID, termExists := student["termId"]
	enrollments, enrollmentsExist := student["enrollments"]

	isClassroomEmpty := !classroomExists || classroomID == ""
	isTermEmpty := !termExists || termID == ""
	isEnrollmentsEmpty := !enrollmentsExist || enrollments == nil

	update := bson.M{
		"$set": bson.M{
			"classroomId": input.ClassroomID,
			"termId":      input.TermID,
			"updatedAt":   time.Now(),
		},
	}

	// If enrollments is empty or null, initialize it with the new enrollment
	if isClassroomEmpty && isTermEmpty && isEnrollmentsEmpty {
		update["$set"].(bson.M)["enrollments"] = []bson.M{
			{
				"classroomId": input.ClassroomID,
				"termId":      input.TermID,
			},
		}
	} else {
		// Otherwise add to existing enrollments
		update["$addToSet"] = bson.M{
			"enrollments": bson.M{
				"classroomId": input.ClassroomID,
				"termId":      input.TermID,
			},
		}
	}

	_, err = col.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student classroom"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student classroom updated"})
}

// UpdateStudent updates student information
func UpdateStudent(c *gin.Context) {
	id := c.Param("id")

	type Input struct {
		FirstName   *string `json:"firstName,omitempty"`
		LastName    *string `json:"lastName,omitempty"`
		Name        *string `json:"name,omitempty"`
		DateOfBirth *string `json:"dob,omitempty"`
		Address     *string `json:"address,omitempty"`
		ProfileURL  *string `json:"profileUrl,omitempty"`
		Sex         *string `json:"sex,omitempty"`
		ParentPhone *string `json:"parentPhone,omitempty"`
		ClassroomID *string `json:"classroomId,omitempty"`
		TermID      *string `json:"termId,omitempty"`
		Status      *string `json:"status,omitempty"`
		Enrollments *[]struct {
			TermID      string `json:"termId,omitempty"`
			ClassroomID string `json:"classroomId,omitempty"`
		} `json:"enrollments,omitempty"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student update input"})
		return
	}

	// Build the update document
	updateDoc := bson.M{}

	if input.FirstName != nil {
		updateDoc["firstName"] = *input.FirstName
	}
	if input.LastName != nil {
		updateDoc["lastName"] = *input.LastName
	}
	if input.Name != nil {
		updateDoc["name"] = *input.Name
	}
	if input.DateOfBirth != nil {
		updateDoc["dob"] = *input.DateOfBirth
	}
	if input.Address != nil {
		updateDoc["address"] = *input.Address
	}
	if input.ProfileURL != nil {
		updateDoc["profileUrl"] = *input.ProfileURL
	}
	if input.Sex != nil {
		updateDoc["sex"] = *input.Sex
	}
	if input.ParentPhone != nil {
		updateDoc["parentPhone"] = *input.ParentPhone
	}
	if input.ClassroomID != nil {
		updateDoc["classroomId"] = *input.ClassroomID
	}
	if input.TermID != nil {
		updateDoc["termId"] = *input.TermID
	}
	if input.Status != nil {
		updateDoc["status"] = *input.Status
	}
	if input.Enrollments != nil {
		// Convert enrollments to bson.M format
		var enrollments []bson.M
		for _, enrollment := range *input.Enrollments {
			enrollmentDoc := bson.M{}
			if enrollment.TermID != "" {
				enrollmentDoc["termId"] = enrollment.TermID
			}
			if enrollment.ClassroomID != "" {
				enrollmentDoc["classroomId"] = enrollment.ClassroomID
			}
			enrollments = append(enrollments, enrollmentDoc)
		}
		updateDoc["enrollments"] = enrollments
	}

	// Add updatedAt timestamp
	updateDoc["updatedAt"] = time.Now()

	// Check if there are any fields to update
	if len(updateDoc) == 1 { // Only updatedAt was added
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("students")

	result, err := col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": updateDoc,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Student updated successfully",
		"id":      id,
	})
}

// GetStudents retrieves all students from the database with optional classroom and term data
// Returns a list of all students as JSON with joined classroom and term information
func GetStudents(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("students")

	// Use aggregation to join with classrooms and terms data
	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{
			"$lookup": bson.M{
				"from":         "classrooms",
				"localField":   "classroomId",
				"foreignField": "_id",
				"as":           "classroom",
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "terms",
				"localField":   "termId",
				"foreignField": "_id",
				"as":           "term",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$classroom",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$term",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
		return
	}
	defer cursor.Close(ctx)

	var students []bson.M
	if err := cursor.All(ctx, &students); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode students"})
		return
	}

	c.JSON(http.StatusOK, students)
}

// get students by classroom id , at case when want to show classroom detail who student has been study at classroom
func GetStudentsByClassroomAndTerm(c *gin.Context) {
	classroom_id := c.Param("classroom_id")
	term_id := c.Param("term_id")

	if classroom_id == "" || term_id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing classroomid or term_id"})
		return
	}

	col := db.MongoClient.Database("school_attendance").Collection("students")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"enrollments": bson.M{
			"$elemMatch": bson.M{
				"classroom_id": classroom_id,
				"term_id":      term_id,
			},
		},
	}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode students"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetStudentsByClassroom fetches students by classroomId only (no termId)
func GetStudentsByClassroom(c *gin.Context) {
	classroomID := c.Param("classroomId")

	if classroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing classroomId"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("students")

	// Find students where classroomId matches
	filter := bson.M{"classroomId": classroomID}

	cursor, err := col.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
		return
	}
	defer cursor.Close(ctx)

	var students []bson.M
	if err := cursor.All(ctx, &students); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode students"})
		return
	}

	c.JSON(http.StatusOK, students)
}
