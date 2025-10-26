package controllers

import (
	"context"

	"net/http"

	"time"

	"school-attendance/db"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
)

func GetSchool(c *gin.Context) {
	schoolID := c.Param("_id")
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School ID is required"})
		return
	}

	col := db.MongoClient.Database("school_attendance").Collection("schools")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var school bson.M
	err := col.FindOne(ctx, bson.M{"_id": schoolID}).Decode(&school)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "School not found"})
		return
	}

	c.JSON(http.StatusOK, school)
}

// CreateSchool creates a new school. Only super_admin users can create schools.
func CreateSchool(c *gin.Context) {
	role := c.GetString("role")
	if role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super_admin can create schools"})
		return
	}

	type SchoolInput struct {
		ID      string `json:"_id"`
		Name    string `json:"name"`
		Address string `json:"address"`
	}

	var input SchoolInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate ID
	if input.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School ID is required"})
		return
	}

	school := bson.M{
		"_id":       input.ID,
		"name":      input.Name,
		"address":   input.Address,
		"createdAt": time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.MongoClient.Database("school_attendance").Collection("schools")
	_, err := collection.InsertOne(ctx, school)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create school"})
		return
	}

	c.JSON(http.StatusCreated, school)
}
