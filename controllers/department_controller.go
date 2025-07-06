package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	// "go.mongodb.org/mongo-driver/mongo"
	"school-attendance/db"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateDepartment(c *gin.Context) {
	type Input struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		SchoolID string `json:"schoolId"`
	}
	id := "department-" + uuid.New().String()
	var input Input
	// if err := c.ShouldBindJSON(&input); err != nil || input.ID == "" || input.SchoolID == "" {
	if err := c.ShouldBindJSON(&input); err != nil || input.SchoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid department input"})
		return
	}

	dept := bson.M{
		"_id":       id,
		"name":      input.Name,
		"schoolId":  input.SchoolID,
		"createdAt": time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("departments")
	_, err := col.InsertOne(ctx, dept)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Department created", "id": id})
}

func GetDepartments(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.MongoClient.Database("school_attendance").Collection("departments")

	cursor, err := collection.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"schoolId": c.GetString("school_id"),
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "schools",
				"localField":   "schoolId",
				"foreignField": "_id",
				"as":           "school",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$school",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch departments"})
		return
	}
	defer cursor.Close(ctx)

	var departments []bson.M
	if err := cursor.All(ctx, &departments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode departments"})
		return
	}

	c.JSON(http.StatusOK, departments)
}

// UpdateDepartment handles updating an existing department's information
// Only admins can update departments, and only for their school
func UpdateDepartment(c *gin.Context) {
	// Get department ID from URL parameter
	departmentID := c.Param("id")
	if departmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department ID is required"})
		return
	}

	// Check if user is admin and get their school ID
	role := c.GetString("role")
	schoolID := c.GetString("school_id")
	if role != "admin" || schoolID == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only school admins can update departments"})
		return
	}

	// Parse input data
	type UpdateInput struct {
		Name string `json:"name"`
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

	// Add updated timestamp
	update["updatedAt"] = time.Now()

	// If nothing to update, return error
	if len(update) <= 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No valid fields to update"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("departments")

	// First check if department exists and belongs to admin's school
	var existingDepartment bson.M
	err := col.FindOne(ctx, bson.M{"_id": departmentID, "schoolId": schoolID}).Decode(&existingDepartment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Department not found or you don't have permission to update it"})
		return
	}

	// Update the department
	_, err = col.UpdateOne(
		ctx,
		bson.M{"_id": departmentID, "schoolId": schoolID},
		bson.M{"$set": update},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update department"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Department updated successfully",
		"id":      departmentID,
	})
}
