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

func CreateTerm(c *gin.Context) {
	type Input struct {
		ID           string `json:"_id"`
		Year         string `json:"year"`
		DepartmentID string `json:"department_id"`
	}
	id := "term-" + uuid.New().String()
	var input Input
	// if err := c.ShouldBindJSON(&input); err != nil || input.ID == "" || input.DepartmentID == "" {
	if err := c.ShouldBindJSON(&input); err != nil || input.DepartmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid term input"})
		return
	}

	term := bson.M{
		"_id":           id,
		"year":          input.Year,
		"department_id": input.DepartmentID,
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

	c.JSON(http.StatusCreated, gin.H{"message": "Term created", "id": input.ID})
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
	fmt.Println(cursor)
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
	fmt.Println(cursor)

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
// Returns the term as JSON or 404 if not found
func GetTerm(c *gin.Context) {
	// Get term ID from URL parameter
	termID := c.Param("id")
	if termID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Term ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")

	// Use aggregation to get term with department information
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

// GetTermsByDepartmentID retrieves all terms by department ID from the database
// Returns the terms as JSON array or 404 if none found
func GetTermsByDepartmentID(c *gin.Context) {
	// Get department ID from URL parameter
	departmentID := c.Param("department_id")
	if departmentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Department ID is required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("terms")

	// Use aggregation to get all terms with department information
	cursor, err := col.Aggregate(ctx, bson.A{
		bson.M{
			"$match": bson.M{
				"department_id": departmentID,
			},
		},
		bson.M{
			"$lookup": bson.M{
				"from":         "terms",
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

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No terms found for the specified department"})
		return
	}

	c.JSON(http.StatusOK, results)
}
