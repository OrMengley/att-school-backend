package controllers

import (
	"context"
	"net/http"
	"time"

	"school-attendance/db"
	"school-attendance/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateGrade(c *gin.Context) {
	type Input struct {
		Name   string `json:"name"`
		TermID string `json:"termId"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil || input.Name == "" || input.TermID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid grade input"})
		return
	}

	schoolID := c.GetString("school_id")
	if schoolID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Missing school ID (only admins can create grades)"})
		return
	}

	grade := models.Grade{
		ID:        "grade-" + uuid.New().String(),
		Name:      input.Name,
		TermID:    input.TermID,
		SchoolID:  schoolID, // ✅ Save schoolId
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("grades")
	_, err := col.InsertOne(ctx, grade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create grade"})
		return
	}

	c.JSON(http.StatusCreated, grade)
}

func GetGrades(c *gin.Context) {
	role := c.GetString("role")
	schoolID := c.GetString("school_id") // available if role == admin

	filter := bson.M{}

	if role == "admin" {
		filter["schoolId"] = schoolID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("grades")
	cursor, err := col.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch grades"})
		return
	}
	defer cursor.Close(ctx)

	var grades []bson.M
	if err := cursor.All(ctx, &grades); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode grades"})
		return
	}

	c.JSON(http.StatusOK, grades)
}

// UpdateGrade handles updating an existing grade's information
// Only admins can update grades, and only for their school
func UpdateGrade(c *gin.Context) {
	// Get grade ID from URL parameter
	gradeID := c.Param("id")
	if gradeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Grade ID is required"})
		return
	}

	// Check if user is admin and get their school ID
	role := c.GetString("role")
	schoolID := c.GetString("school_id")
	if role != "admin" || schoolID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only school admins can update grades"})
		return
	}

	// Parse input data
	type UpdateInput struct {
		Name   string `json:"name"`
		TermID string `json:"termId"`
	}

	var input UpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid update data"})
		return
	}

	// Create update document
	update := bson.M{}

	if input.Name != "" {
		update["name"] = input.Name
	}

	if input.TermID != "" {
		update["termId"] = input.TermID
	}

	// Add updated timestamp
	update["updatedAt"] = time.Now()

	// If nothing to update, return error
	if len(update) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("grades")

	// First check if grade exists and belongs to admin's school
	var existingGrade models.Grade
	err := col.FindOne(ctx, bson.M{"_id": gradeID, "schoolId": schoolID}).Decode(&existingGrade)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grade not found or you don't have permission to update it"})
		return
	}

	// Update the grade
	_, err = col.UpdateOne(
		ctx,
		bson.M{"_id": gradeID, "schoolId": schoolID},
		bson.M{"$set": update},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update grade"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Grade updated successfully",
		"id":      gradeID,
	})
}
