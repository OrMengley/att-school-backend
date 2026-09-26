package controllers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"school-attendance/db"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateTerm(c *gin.Context) {
	type Input struct {
		ID           string `json:"_id"`
		Year         string `json:"year"`
		DepartmentID string `json:"department_id"`
		DepartmentId string `json:"departmentId"`
	}
	var input Input
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid term input"})
		return
	}

	deptID := strings.TrimSpace(input.DepartmentID)
	if deptID == "" {
		deptID = strings.TrimSpace(input.DepartmentId)
	}

	if deptID == "" || strings.TrimSpace(input.Year) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Year and Department ID are required"})
		return
	}

	id := strings.TrimSpace(input.ID)
	if id == "" {
		id = "term-" + uuid.New().String()
	}

	term := bson.M{
		"_id":           id,
		"year":          strings.TrimSpace(input.Year),
		"department_id": deptID,
		"deleted":       false,
		"createdAt":     time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")

	_, err := col.InsertOne(ctx, term)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Insert failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Term created", "id": id})
}

func UpdateTerm(c *gin.Context) {
	termID := c.Param("id")
	if termID == "" {
		termID = c.Param("_id")
	}
	if termID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term ID is required"})
		return
	}

	var input struct {
		Year         *string `json:"year,omitempty"`
		DepartmentID *string `json:"department_id,omitempty"`
		DepartmentId *string `json:"departmentId,omitempty"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	update := bson.M{}
	if input.Year != nil {
		update["year"] = strings.TrimSpace(*input.Year)
	}
	if input.DepartmentID != nil {
		update["department_id"] = strings.TrimSpace(*input.DepartmentID)
	} else if input.DepartmentId != nil {
		update["department_id"] = strings.TrimSpace(*input.DepartmentId)
	}

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	update["updatedAt"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	res, err := col.UpdateOne(ctx, bson.M{"_id": termID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update term"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Term not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Term updated successfully"})
}

func GetTermsBySchoolID(c *gin.Context) {
	schoolID := c.Query("school_id")
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"deleted": bson.M{"$ne": true},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "departments",
				"localField":   "department_id",
				"foreignField": "_id",
				"as":           "department",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$department",
				"preserveNullAndEmptyArrays": true,
			},
		},
		bson.M{
			"$match": bson.M{
				"department.school_id": schoolID,
			},
		},
	})

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

func GetTerms(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.MongoClient.Database("school_attendance").Collection("terms")

	cursor, err := collection.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"deleted": bson.M{"$ne": true},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "departments",
				"localField":   "department_id",
				"foreignField": "_id",
				"as":           "department",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$department",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})

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

// GetTerm retrieves a single term by ID from the database
func GetTerm(c *gin.Context) {
	termID := c.Param("id")
	if termID == "" {
		termID = c.Param("_id")
	}
	if termID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")

	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"_id": termID,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "departments",
				"localField":   "department_id",
				"foreignField": "_id",
				"as":           "department",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$department",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch term"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode term"})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Term not found"})
		return
	}

	c.JSON(http.StatusOK, results[0])
}

// GetTermsByDepartmentID retrieves all active terms by department ID
func GetTermsByDepartmentID(c *gin.Context) {
	departmentID := c.Param("department_id")
	if departmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")

	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"department_id": departmentID,
				"deleted":       bson.M{"$ne": true},
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "departments",
				"localField":   "department_id",
				"foreignField": "_id",
				"as":           "department",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$department",
				"preserveNullAndEmptyArrays": true,
			},
		},
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch terms"})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode terms"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// GetTrashedTermsBySchoolID retrieves trashed (soft-deleted) terms for a school
func GetTrashedTermsBySchoolID(c *gin.Context) {
	schoolID := c.Query("school_id")
	if schoolID == "" {
		schoolID = c.GetString("school_id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")

	pipeline := bson.A{
		bson.M{
			"$match": bson.M{
				"deleted": true,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "departments",
				"localField":   "department_id",
				"foreignField": "_id",
				"as":           "department",
			},
		},
		bson.M{
			"$unwind": bson.M{
				"path":                       "$department",
				"preserveNullAndEmptyArrays": true,
			},
		},
	}

	if schoolID != "" {
		pipeline = append(pipeline, bson.M{
			"$match": bson.M{
				"department.school_id": schoolID,
			},
		})
	}

	cursor, err := col.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trashed terms"})
		return
	}
	defer cursor.Close(ctx)

	var terms []bson.M
	if err := cursor.All(ctx, &terms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode trashed terms"})
		return
	}

	c.JSON(http.StatusOK, terms)
}

// SoftDeleteTerm moves a single term to trash
func SoftDeleteTerm(c *gin.Context) {
	termID := c.Param("id")
	if termID == "" {
		termID = c.Param("_id")
	}
	if termID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	now := time.Now()
	res, err := col.UpdateOne(ctx, bson.M{"_id": termID}, bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": now,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move term to trash"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Term not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Term moved to trash successfully", "id": termID})
}

// RestoreTerm restores a trashed term back to active
func RestoreTerm(c *gin.Context) {
	termID := c.Param("id")
	if termID == "" {
		termID = c.Param("_id")
	}
	if termID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	res, err := col.UpdateOne(ctx, bson.M{"_id": termID}, bson.M{
		"$set": bson.M{
			"deleted": false,
		},
		"$unset": bson.M{
			"deletedAt": "",
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore term"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Term not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Term restored successfully", "id": termID})
}

// PermanentDeleteTerm permanently deletes a term from MongoDB
func PermanentDeleteTerm(c *gin.Context) {
	termID := c.Param("id")
	if termID == "" {
		termID = c.Param("_id")
	}
	if termID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	res, err := col.DeleteOne(ctx, bson.M{"_id": termID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete term permanently"})
		return
	}
	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Term not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Term deleted permanently", "id": termID})
}

// BulkSoftDeleteTerms moves multiple terms to trash
func BulkSoftDeleteTerms(c *gin.Context) {
	var input struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || len(input.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List of term IDs is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	now := time.Now()
	res, err := col.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": input.IDs}}, bson.M{
		"$set": bson.M{
			"deleted":   true,
			"deletedAt": now,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move terms to trash"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Terms moved to trash",
		"count":   res.ModifiedCount,
	})
}

// BulkRestoreTerms restores multiple trashed terms
func BulkRestoreTerms(c *gin.Context) {
	var input struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || len(input.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List of term IDs is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	res, err := col.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": input.IDs}}, bson.M{
		"$set": bson.M{
			"deleted": false,
		},
		"$unset": bson.M{
			"deletedAt": "",
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore terms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Terms restored successfully",
		"count":   res.ModifiedCount,
	})
}

// BulkPermanentDeleteTerms permanently deletes multiple terms
func BulkPermanentDeleteTerms(c *gin.Context) {
	var input struct {
		IDs []string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || len(input.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "List of term IDs is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")
	res, err := col.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": input.IDs}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to permanently delete terms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Terms deleted permanently",
		"count":   res.DeletedCount,
	})
}
