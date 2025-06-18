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

func CreateTerm(c *gin.Context) {
	type Input struct {
		ID           string `json:"id"`
		Year         string `json:"year"`
		DepartmentID string `json:"departmentId"`
	}
	id := "term-" + uuid.New().String()
	var input Input
	// if err := c.ShouldBindJSON(&input); err != nil || input.ID == "" || input.DepartmentID == "" {
	if err := c.ShouldBindJSON(&input); err != nil || input.DepartmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid term input"})
		return
	}

	term := bson.M{
		"_id":          id,
		"year":         input.Year,
		"departmentId": input.DepartmentID,
		"createdAt":    time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	_, err := col.InsertOne(ctx, term)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Term created", "id": input.ID})
}

func GetTerms(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.MongoClient.Database("school_attendance").Collection("terms")

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch terms"})
		return
	}
	defer cursor.Close(ctx)

	var terms []bson.M
	if err := cursor.All(ctx, &terms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode terms"})
		return
	}

	c.JSON(http.StatusOK, terms)
}
