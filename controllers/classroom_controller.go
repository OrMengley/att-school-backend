package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"school-attendance/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // ✅ new import
	"go.mongodb.org/mongo-driver/bson"
)

// CreateClassroom handles the creation of a new classroom
// It accepts a JSON payload with name and gradeId
// Returns the newly created classroom ID on success
func CreateClassroom(c *gin.Context) {
	type Input struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		GradeID string `json:"gradeId"`
		TermID  string `json:"termId"`
	}

	var input Input
	// if err := c.ShouldBindJSON(&input); err != nil || input.ID == "" || input.GradeID == "" {
	if err := c.ShouldBindJSON(&input); err != nil || input.GradeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom input"})
		return
	}
	id := "classroom-" + uuid.New().String() // ✅ Auto UUID
	classroom := bson.M{
		"_id":       id,
		"name":      input.Name,
		"gradeId":   input.GradeID,
		"termId":    input.TermID,
		"createdAt": time.Now(),
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
				"localField":   "gradeId",
				"foreignField": "_id",
				"as":           "grade",
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

// UpdateClassroom handles updating an existing classroom
// It accepts a JSON payload with classroom ID and updated fields
// Returns success message on successful update
func UpdateClassroom(c *gin.Context) {
	type Input struct {
		Name    string `json:"name"`
		GradeID string `json:"gradeId"`
		TermID  string `json:"termId"`
	}

	// Get classroom ID from URL parameter
	classroomID := c.Param("id")
	if classroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Classroom ID is required"})
		return
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom input"})
		return
	}

	// Build update fields
	updateFields := bson.M{
		"updatedAt": time.Now(),
	}

	if input.Name != "" {
		updateFields["name"] = input.Name
	}
	if input.GradeID != "" {
		updateFields["gradeId"] = input.GradeID
	}
	if input.TermID != "" {
		updateFields["termId"] = input.TermID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	// Update the classroom
	result, err := col.UpdateOne(
		ctx,
		bson.M{"_id": classroomID},
		bson.M{"$set": updateFields},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update classroom"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Classroom not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Classroom updated successfully"})
}
