package controllers

import (
	"context"

	"net/http"

	"time"

	"school-attendance/db"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
)

func GetSchool(c *gin.Context) {
	schoolID := c.Param("id")
	if schoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "School ID is required"})
		return
	}

	col := db.MongoClient.Database("school_attendance").Collection("schools")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var school bson.M
	err := col.FindOne(ctx, bson.M{"_id": schoolID}).Decode(&school)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "School not found"})
		return
	}

	c.JSON(http.StatusOK, school)
}
