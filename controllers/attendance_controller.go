package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"school-attendance/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func RecordAttendance(c *gin.Context) {
	type Input struct {
		StudentID string `json:"studentId"`
		Status    string `json:"status"` // "checkin" or "checkout"
	}

	var input Input
	if err := c.ShouldBindJSON(&input); err != nil || input.StudentID == "" || (input.Status != "checkin" && input.Status != "checkout") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1️⃣ Fetch student to get classroomId and parentPhone
	studentsCol := db.MongoClient.Database("school_attendance").Collection("students")
	var student bson.M
	err := studentsCol.FindOne(ctx, bson.M{"_id": input.StudentID}).Decode(&student)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// 2️⃣ Prepare attendance entry
	id := "attendance-" + uuid.New().String()
	attendance := bson.M{
		"_id":         id,
		"studentId":   input.StudentID,
		"classroomId": student["classroomId"],
		"status":      input.Status,
		"timestamp":   time.Now(),
	}

	// 3️⃣ Insert attendance
	attCol := db.MongoClient.Database("school_attendance").Collection("attendance")
	_, err = attCol.InsertOne(ctx, attendance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record attendance"})
		return
	}

	// 4️⃣ Log notification
	parentPhone := student["parentPhone"]
	fmt.Printf("📢 Notifying parent %s: student %s has %s\n", parentPhone, student["name"], input.Status)

	c.JSON(http.StatusCreated, gin.H{"message": "Attendance recorded", "id": id})
}
