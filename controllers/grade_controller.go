package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/google/uuid"
	"school-attendance/db"
)

func CreateGrade(c *gin.Context) {
	type Input struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		TermID string `json:"termId"`
	}
	id := "grade-" + uuid.New().String()
	var input Input
	// if err := c.ShouldBindJSON(&input); err != nil || input.ID == "" || input.TermID == "" {
	if err := c.ShouldBindJSON(&input); err != nil || input.TermID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid grade input"})
		return
	}

	grade := bson.M{
		"_id":       id,
		"name":      input.Name,
		"termId":    input.TermID,
		"createdAt": time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("grades")
	_, err := col.InsertOne(ctx, grade)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Grade created", "id": input.ID})
}
