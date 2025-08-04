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

// CheckInOrOutStudent marks a student as checked in or out
func CheckInOrOutStudent(c *gin.Context) {
	var input struct {
		StudentID   string `json:"studentId"`
		ClassroomID string `json:"classroomId"`
		TeacherID   string `json:"teacherId"`
		Status      string `json:"status"` // "checkin" or "checkout"
	}

	if err := c.ShouldBindJSON(&input); err != nil ||
		input.StudentID == "" || input.ClassroomID == "" || input.TeacherID == "" ||
		(input.Status != "checkin" && input.Status != "checkout") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("attendance")

	today := time.Now().Format("2006-01-02")

	// Check if an attendance record exists for this student/classroom/today
	filter := bson.M{
		"studentId":   input.StudentID,
		"classroomId": input.ClassroomID,
		"date":        today,
	}

	var existing models.Attendance
	err := col.FindOne(ctx, filter).Decode(&existing)

	if err == nil {
		// Update the existing record's status and time
		update := bson.M{
			"$set": bson.M{
				"status":    input.Status,
				"time":      time.Now(),
				"teacherId": input.TeacherID,
			},
		}
		_, _ = col.UpdateOne(ctx, filter, update)

		c.JSON(http.StatusOK, gin.H{
			"message": "Attendance updated",
			"status":  input.Status,
		})
		return
	}

	// Otherwise, insert a new attendance record
	attendance := models.Attendance{
		ID:          uuid.New().String(),
		StudentID:   input.StudentID,
		ClassroomID: input.ClassroomID,
		Date:        today,
		Status:      input.Status,
		Time:        time.Now(),
		TeacherID:   input.TeacherID,
	}

	_, errInsert := col.InsertOne(ctx, attendance)
	if errInsert != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Attendance marked",
		"status":  input.Status,
	})
}
