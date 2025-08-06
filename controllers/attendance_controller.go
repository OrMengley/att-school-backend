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

// CheckInOrOutStudent handles check-in or check-out for a student
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

	filter := bson.M{
		"studentId":   input.StudentID,
		"classroomId": input.ClassroomID,
		"date":        today,
	}

	var existing models.Attendance
	err := col.FindOne(ctx, filter).Decode(&existing)

	update := bson.M{}
	now := time.Now()

	if err == nil {
		// Record already exists → Update status
		if input.Status == "checkin" {
			update = bson.M{
				"$set": bson.M{
					"checkinTime": now,
					"teacherId":   input.TeacherID,
					"status":      "checkin",
				},
			}
		} else {
			update = bson.M{
				"$set": bson.M{
					"checkoutTime": now,
					"teacherId":    input.TeacherID,
					"status":       "checkout",
				},
			}
		}

		_, err = col.UpdateOne(ctx, filter, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update attendance"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Attendance updated",
			"status":  input.Status,
		})
		return
	}

	// No existing record → Create new check-in or check-out
	newAttendance := models.Attendance{
		ID:          "att-" + uuid.New().String(),
		StudentID:   input.StudentID,
		ClassroomID: input.ClassroomID,
		TeacherID:   input.TeacherID,
		Date:        today,
		Status:      input.Status,
	}

	if input.Status == "checkin" {
		newAttendance.CheckinTime = now
	} else {
		newAttendance.CheckoutTime = now
	}

	_, err = col.InsertOne(ctx, newAttendance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Attendance recorded",
		"status":  input.Status,
	})
}
