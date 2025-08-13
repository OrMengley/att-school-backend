// controllers/parent_controller.go
package controllers

import (
	"context"
	"net/http"
	"time"

	"school-attendance/db"
	"school-attendance/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// GET /parents/me/children
// Returns: { "children": [...], "subscriptionMap": { "<studentId>": <unixSec>, ... } }
// controllers/parent_controller.go
func GetMyChildren(c *gin.Context) {
	if c.GetString("role") != "parent" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only parent can view their children"})
		return
	}
	parentID := c.GetString("user_id") // <-- match your AuthMiddleware key
	if parentID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCol := db.MongoClient.Database("school_attendance").Collection("users")
	var parent bson.M
	if err := usersCol.FindOne(ctx, bson.M{"_id": parentID, "role": "parent"}).Decode(&parent); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Parent not found"})
		return
	}

	// subscribedStudents: []string
	idsAny, _ := parent["subscribedStudents"].(bson.A)
	if len(idsAny) == 0 {
		c.JSON(http.StatusOK, gin.H{"children": []any{}, "subscriptionMap": gin.H{}})
		return
	}

	ids := make([]string, 0, len(idsAny))
	for _, v := range idsAny {
		if s, ok := v.(string); ok {
			ids = append(ids, s)
		}
	}

	studentsCol := db.MongoClient.Database("school_attendance").Collection("students")
	cur, err := studentsCol.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
		return
	}
	defer cur.Close(ctx)

	var children []bson.M
	if err := cur.All(ctx, &children); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode students"})
		return
	}

	// optional: include subscriptionMap if you store it on user
	subMap, _ := parent["subscriptionMap"].(bson.M)
	c.JSON(http.StatusOK, gin.H{"children": children, "subscriptionMap": subMap})
}

// GET /attendance/student/:studentId?from=YYYY-MM-DD&to=YYYY-MM-DD
// Parents can only read attendance for students they’re subscribed to.
func GetAttendanceByStudent(c *gin.Context) {
	role := c.GetString("role")
	userID := c.GetString("user_id")
	studentID := c.Param("studentId")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing studentId"})
		return
	}

	// If parent, verify the studentId is in their subscribedStudents
	if role == "parent" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var parent models.User
		if err := db.MongoClient.Database("school_attendance").
			Collection("users").
			FindOne(ctx, bson.M{"_id": userID, "role": "parent"}).
			Decode(&parent); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		allowed := false
		for _, id := range parent.SubscribedStudents {
			if id == studentID {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not your child"})
			return
		}
	}

	from := c.Query("from")
	to := c.Query("to")

	// Build filter
	filter := bson.M{"studentId": studentID}
	if from != "" || to != "" {
		dateCond := bson.M{}
		if from != "" {
			dateCond["$gte"] = from
		}
		if to != "" {
			dateCond["$lte"] = to
		}
		filter["date"] = dateCond
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("attendance")
	cur, err := col.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch attendance"})
		return
	}
	defer cur.Close(ctx)

	var out []models.Attendance
	if err := cur.All(ctx, &out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode attendance"})
		return
	}

	c.JSON(http.StatusOK, out)
}

// controllers/parent_controller.go

// GET /parents/me/children/:studentId/attendance?from=YYYY-MM-DD&to=YYYY-MM-DD
func GetMyChildAttendance(c *gin.Context) {
	role := c.GetString("role")
	if role != "parent" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only parent can view attendance"})
		return
	}
	parentID := c.GetString("user_id") // make sure your AuthMiddleware sets "user_id"
	studentID := c.Param("studentId")
	if parentID == "" || studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parent or student"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dbx := db.MongoClient.Database("school_attendance")

	// 1) Verify the student is in this parent's subscribedStudents
	var parentDoc bson.M
	if err := dbx.Collection("users").
		FindOne(ctx, bson.M{"_id": parentID, "role": "parent"}).
		Decode(&parentDoc); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Parent not found"})
		return
	}
	ok := false
	if arr, _ := parentDoc["subscribedStudents"].(bson.A); len(arr) > 0 {
		for _, v := range arr {
			if s, _ := v.(string); s == studentID {
				ok = true
				break
			}
		}
	}
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not subscribed to this student"})
		return
	}

	// 2) Fetch attendance for optional range
	from := c.Query("from")
	to := c.Query("to")

	filter := bson.M{"studentId": studentID}
	if from != "" || to != "" {
		and := bson.A{bson.M{"studentId": studentID}}
		if from != "" {
			and = append(and, bson.M{"date": bson.M{"$gte": from}})
		}
		if to != "" {
			and = append(and, bson.M{"date": bson.M{"$lte": to}})
		}
		filter = bson.M{"$and": and}
	}

	cur, err := dbx.Collection("attendance").Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch attendance"})
		return
	}
	defer cur.Close(ctx)

	var out []models.Attendance
	if err := cur.All(ctx, &out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode attendance"})
		return
	}

	c.JSON(http.StatusOK, out)
}
