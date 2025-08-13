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
		SessionID   string `json:"sessionId"`
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
		"sessionId":   input.SessionID,
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

// GetAttendanceBySession returns all attendance records for a given session
func GetAttendanceBySession(c *gin.Context) {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing sessionId"})
		return
	}

	collection := db.MongoClient.Database("school_attendance").Collection("attendance")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Filter by sessionId
	filter := bson.M{"sessionId": sessionId}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attendance"})
		return
	}
	defer cursor.Close(ctx)

	var attendances []models.Attendance
	if err := cursor.All(ctx, &attendances); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode attendance"})
		return
	}

	// Optionally trim the response if only ID and status is needed:
	trimmed := make([]gin.H, 0, len(attendances))
	for _, a := range attendances {
		trimmed = append(trimmed, gin.H{
			"studentId": a.StudentID,
			"status":    a.Status,
		})
	}

	c.JSON(http.StatusOK, trimmed)
}
