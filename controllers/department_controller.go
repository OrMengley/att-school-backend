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

	cursor, err := collection.Find(ctx, bson.M{})
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
