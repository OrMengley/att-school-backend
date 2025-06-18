package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/google/uuid" // ✅ new import
	"school-attendance/db"
)

// CreateClassroom handles the creation of a new classroom
// It accepts a JSON payload with name and gradeId
// Returns the newly created classroom ID on success
func CreateClassroom(c *gin.Context) {
	type Input struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		GradeID string `json:"gradeId"`
	}

	var input Input
	// if err := c.ShouldBindJSON(&input); err != nil || input.ID == "" || input.GradeID == "" {
	if err := c.ShouldBindJSON(&input); err != nil || input.GradeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom input"})
		return
	}
	id := "classroom-" + uuid.New().String() // ✅ Auto UUID
	classroom := bson.M{
		"_id":      id,
		"name":     input.Name,
		"gradeId":  input.GradeID,
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
	col := db.MongoClient.Database("school_attendance").Collection("classrooms")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, bson.M{})
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