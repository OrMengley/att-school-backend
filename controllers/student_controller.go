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
		Name        string `json:"name"`
		ClassroomID string `json:"classroomId"`
		ParentPhone string `json:"parentPhone"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil || input.ClassroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid student input"})
		return
	}

	id := "student-" + uuid.New().String()

	student := bson.M{
		"_id":         id,
		"name":        input.Name,
		"classroomId": input.ClassroomID,
		"parentPhone": input.ParentPhone,
		"createdAt":   time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("students")
	_, err := col.InsertOne(ctx, student)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create student"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Student created", "id": id})
}

func GetStudents(c *gin.Context) {
	col := db.MongoClient.Database("school_attendance").Collection("students")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := col.Find(ctx, bson.M{})
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

func UpdateStudentClassroom(c *gin.Context) {
	id := c.Param("id")

	type Input struct {
		ClassroomID string `json:"classroomId"`
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil || input.ClassroomID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid classroom update input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("students")

	_, err := col.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"classroomId": input.ClassroomID},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update student classroom"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Student classroom updated"})
}
