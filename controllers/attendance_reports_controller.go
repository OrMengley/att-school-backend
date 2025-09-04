package controllers

import (
	"context"
	"net/http"
	"time"

	"school-attendance/db"
	"school-attendance/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ---- helpers ----

func parseDateRange(c *gin.Context) (from, to string) {
	from = c.Query("from")
	to = c.Query("to")
	return
}

func addDateFilter(base bson.M, date, from, to string) bson.M {
	if date != "" {
		base["date"] = date
		return base
	}
	if from != "" || to != "" {
		and := bson.A{bson.M{"date": bson.M{"$exists": true}}}
		if from != "" {
			and = append(and, bson.M{"date": bson.M{"$gte": from}})
		}
		if to != "" {
			and = append(and, bson.M{"date": bson.M{"$lte": to}})
		}
		return bson.M{"$and": andAppend(base, and)}
	}
	return base
}

func andAppend(base bson.M, and bson.A) bson.A {
	// merge existing base conditions into $and
	for k, v := range base {
		and = append(and, bson.M{k: v})
	}
	return and
}

func attendanceCol() *mongo.Collection {
	return db.MongoClient.Database("school_attendance").Collection("attendance")
}

func classroomsCol() *mongo.Collection {
	return db.MongoClient.Database("school_attendance").Collection("classrooms")
}

// fetch classroom IDs that belong to this teacher
func getTeacherClassroomIDs(ctx context.Context, teacherID string) ([]string, error) {
	cur, err := classroomsCol().Find(ctx, bson.M{"teacher_id": teacherID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	type row struct {
		ID string `bson:"_id"`
	}
	var out []row
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(out))
	for _, r := range out {
		ids = append(ids, r.ID)
	}
	return ids, nil
}

// ---- endpoints ----

// GET /attendance/by-classroom/:classroomId?date=YYYY-MM-DD&from=YYYY-MM-DD&to=YYYY-MM-DD
func AttendanceByClassroom(c *gin.Context) {
	classroom_id := c.Param("classroom_id")
	if classroom_id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing classroom_id"})
		return
	}

	role := c.GetString("role")
	user_id := c.GetString("user_id")
	school_id := c.GetString("school_id")

	// ---- permission checks ----
	usersCol := db.MongoClient.Database("school_attendance").Collection("users")
	classroomsCol := db.MongoClient.Database("school_attendance").Collection("classrooms")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Load classroom to check its school
	var classroom bson.M
	if err := classroomsCol.FindOne(ctx, bson.M{"_id": classroom_id}).Decode(&classroom); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "classroom not found"})
		return
	}
	clsSchoolId, _ := classroom["school_id"].(string)

	switch role {
	case "super_admin":
		// ok
	case "admin":
		if school_id == "" || clsSchoolId == "" || school_id != clsSchoolId {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: different school"})
			return
		}
	case "teacher":
		// verify this teacher owns the classroom
		var teacher bson.M
		if err := usersCol.FindOne(ctx, bson.M{"_id": user_id, "role": "teacher"}).Decode(&teacher); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		// classroomIds: []string
		okOwned := false
		if arr, ok := teacher["classroomIds"].(bson.A); ok {
			for _, v := range arr {
				if s, _ := v.(string); s == classroom_id {
					okOwned = true
					break
				}
			}
		}
		if !okOwned {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden: not your classroom"})
			return
		}
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	// ---- fetch today's attendance for the classroom ----
	attendanceCol := db.MongoClient.Database("school_attendance").Collection("attendance")
	today := time.Now().Format("2006-01-02")

	filter := bson.M{
		"classroom_id": classroom_id,
		"date":         today,
	}

	cur, err := attendanceCol.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query attendance"})
		return
	}
	defer cur.Close(ctx)

	var out []bson.M
	if err := cur.All(ctx, &out); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decode attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"classroom_id": classroom_id,
		"date":         today,
		"items":        out,
	})
}

// GET /attendance/student/:studentId?from=&to=
func AttendanceByStudent(c *gin.Context) {
	role := c.GetString("role")
	if role != "teacher" && role != "parent" && role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	studentID := c.Param("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing studentId"})
		return
	}
	from, to := parseDateRange(c)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	filter := bson.M{"student_id": studentID}
	filter = addDateFilter(filter, "", from, to)

	cur, err := attendanceCol().Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer cur.Close(ctx)

	var rows []models.Attendance
	if err := cur.All(ctx, &rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode failed"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// GET /attendance/range?from=&to=   (teacher-scoped across all their classrooms)
func AttendanceRangeForTeacher(c *gin.Context) {
	role := c.GetString("role")
	if role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "teacher only"})
		return
	}
	from, to := parseDateRange(c)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	teacherID := c.GetString("user_id")
	classroomIDs, err := getTeacherClassroomIDs(ctx, teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve classrooms"})
		return
	}
	if len(classroomIDs) == 0 {
		c.JSON(http.StatusOK, []models.Attendance{})
		return
	}

	filter := bson.M{"classroom": bson.M{"$in": classroomIDs}}
	filter = addDateFilter(filter, "", from, to)

	cur, err := attendanceCol().Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer cur.Close(ctx)

	var rows []models.Attendance
	if err := cur.All(ctx, &rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode failed"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// GET /attendance/today   (teacher-scoped, today only)
func AttendanceTodayForTeacher(c *gin.Context) {
	role := c.GetString("role")
	if role != "teacher" {
		c.JSON(http.StatusForbidden, gin.H{"error": "teacher only"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	teacherID := c.GetString("user_id")
	classroomIDs, err := getTeacherClassroomIDs(ctx, teacherID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve classrooms"})
		return
	}
	if len(classroomIDs) == 0 {
		c.JSON(http.StatusOK, []models.Attendance{})
		return
	}

	today := time.Now().Format("2006-01-02")
	filter := bson.M{
		"classroom_id": bson.M{"$in": classroomIDs},
		"date":         today,
	}

	cur, err := attendanceCol().Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer cur.Close(ctx)

	var rows []models.Attendance
	if err := cur.All(ctx, &rows); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode failed"})
		return
	}
	c.JSON(http.StatusOK, rows)
}
