// controllers/user_controller.go
package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"school-attendance/db"
	"school-attendance/models"
	"school-attendance/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// sanitizeUser removes sensitive fields before sending to client
func sanitizeUser(u models.User) gin.H {
	return gin.H{
		"_id":       u.ID,
		"name":      u.Name,
		"email":     u.Email,
		"role":      u.Role,
		"status":    u.Status,
		"school_id": u.SchoolID,
		// include these if you want them client-side
		"classroomIds":       u.ClassroomIDs,
		"subscribedStudents": u.SubscribedStudents,
	}
}

func GetAllUsers(c *gin.Context) {
	collection := db.MongoClient.Database("school_attendance").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Read current user info from JWT
	currentRole := c.GetString("role")
	schoolID := c.GetString("school_id")

	// Read optional ?role query
	queryRole := c.Query("role")

	// Build filter
	filter := bson.M{}
	if queryRole != "" {
		filter["role"] = queryRole
	}

	// If current user is a school admin, only allow access to their school's users
	if currentRole == "admin" {
		filter["school_id"] = schoolID
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Println("Failed to fetch users:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	if user.ID == "" {
		user.ID = primitive.NewObjectID().Hex()
	}

	if user.Role == "admin" || user.Role == "teacher" || user.Role == "parent" {
		if user.SchoolID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "School ID is required for this role"})
			return
		}
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	user.Password = hashedPassword

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.MongoClient.Database("school_attendance").Collection("users")
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, sanitizeUser(user))
}

// func CreateUser(c *gin.Context) {
// 	var user models.User

// 	if err := c.ShouldBindJSON(&user); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	if user.Password == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
// 		return
// 	}

// 	if user.ID == "" {
// 		user.ID = primitive.NewObjectID().Hex()
// 	}

// 	if user.Role == "admin" || user.Role == "teacher" || user.Role == "parent" {
// 		if user.SchoolID == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "School ID is required for this role"})
// 			return
// 		}
// 	}

// 	hashedPassword, err := utils.HashPassword(user.Password)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
// 		return
// 	}
// 	user.Password = hashedPassword

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	collection := db.MongoClient.Database("school_attendance").Collection("users")
// 	_, err = collection.InsertOne(ctx, user)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"id":       user.ID,
// 		"name":     user.Name,
// 		"email":    user.Email,
// 		"role":     user.Role,
// 		"schoolId": user.SchoolID,
// 	})

// 	fmt.Println("📦 DEBUG incoming password:", user.Password)
// }

func Login(c *gin.Context) {
	type LoginInput struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.Email = strings.TrimSpace(input.Email)
	input.Password = strings.TrimSpace(input.Password)

	collection := db.MongoClient.Database("school_attendance").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	if !utils.CheckPasswordHash(input.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check user status if it exists
	if user.Status != "" && user.Status != "active" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is inactive"})
		return
	}

	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	if user.SchoolID != "" {
		claims["school_id"] = user.SchoolID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("SECRET_KEY"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  sanitizeUser(user),
	})
}

// func Login(c *gin.Context) {

// 	type LoginInput struct {
// 		Email    string `json:"email"`
// 		Password string `json:"password"`
// 	}
// 	var input LoginInput
// 	if err := c.ShouldBindJSON(&input); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // return full error
// 		return
// 	}

// 	// Trim both email and password to prevent login issues due to whitespace
// 	input.Email = strings.TrimSpace(input.Email)
// 	input.Password = strings.TrimSpace(input.Password)

// 	collection := db.MongoClient.Database("school_attendance").Collection("users")

// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	var user models.User
// 	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
// 	if err != nil {
// 		// Security best practice: Don't reveal whether email or password is wrong
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
// 		return
// 	}

// 	match := utils.CheckPasswordHash(input.Password, user.Password)

// 	if !match {
// 		// Security best practice: Use the same message as email not found
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
// 		return
// 	}

// 	claims := jwt.MapClaims{
// 		"user_id": user.ID,
// 		"role":    user.Role,
// 		"exp":     time.Now().Add(time.Hour * 24).Unix(),
// 	}

// 	if user.SchoolID != "" {
// 		claims["school_id"] = user.SchoolID
// 	}

// 	// Generate JWT token
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	tokenString, err := token.SignedString([]byte("SECRET_KEY"))

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"token": tokenString, "user": user})
// }

// AdminUpdateUser allows admin/super_admin to update basic user fields
func AdminUpdateUser(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin/super_admin can edit users"})
		return
	}

	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user id"})
		return
	}

	var input struct {
		Name     *string `json:"name,omitempty"`
		Email    *string `json:"email,omitempty"`
		Status   *string `json:"status,omitempty"`
		Role     *string `json:"role,omitempty"`      // "super_admin" | "admin" | "teacher" | "parent"
		SchoolID *string `json:"school_id,omitempty"` // required for admin/teacher/parent
		// ClassroomIDs *[]string `json:"classroomIds,omitempty"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	update := bson.M{}
	if input.Name != nil {
		update["name"] = strings.TrimSpace(*input.Name)
	}
	if input.Email != nil {
		update["email"] = strings.TrimSpace(*input.Email)
	}
	if input.Role != nil {
		update["role"] = strings.TrimSpace(*input.Role)
	}
	if input.Status != nil {
		update["status"] = strings.TrimSpace(*input.Status)
	}
	if input.SchoolID != nil {
		update["school_id"] = strings.TrimSpace(*input.SchoolID)
	}
	// if input.ClassroomIDs != nil {
	// 	update["classroomIds"] = *input.ClassroomIDs
	// }

	if len(update) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}
	update["updatedAt"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("users")
	res, err := col.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": update})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Return updated (sanitized) user
	var updated models.User
	if err := col.FindOne(ctx, bson.M{"_id": userID}).Decode(&updated); err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "User updated", "user": sanitizeUser(updated)})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

// AdminSetUserPassword lets admin/super_admin reset a user's password
func AdminSetUserPassword(c *gin.Context) {
	role := c.GetString("role")
	if role != "admin" && role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin/super_admin can reset passwords"})
		return
	}

	userID := c.Param("_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user id"})
		return
	}

	var input struct {
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.NewPassword) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid password payload"})
		return
	}

	hashed, err := utils.HashPassword(strings.TrimSpace(input.NewPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	col := db.MongoClient.Database("school_attendance").Collection("users")
	res, err := col.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{
		"$set": bson.M{
			"password":  hashed,
			"updatedAt": time.Now(),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set password"})
		return
	}
	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// TestLoginEndpoint creates a test user and attempts to log in with it
// This is a special endpoint for debugging the login process
func TestLoginEndpoint(c *gin.Context) {
	// Step 1: Create a test user
	testUser := models.User{
		ID:       "test_" + time.Now().Format("20060102150405"),
		Name:     "Test User",
		Email:    "test_" + time.Now().Format("20060102150405") + "@example.com",
		Password: "testpassword123",
		Role:     "user",
	}

	log.Println("DEBUG: Creating test user with ID:", testUser.ID)
	log.Println("DEBUG: Test user email:", testUser.Email)
	log.Println("DEBUG: Test user password:", testUser.Password)

	// Hash the password
	hashedPassword, err := utils.HashPassword(testUser.Password)
	if err != nil {
		log.Println("DEBUG: Failed to hash password:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	log.Println("DEBUG: Hashed password:", hashedPassword)
	testUser.Password = hashedPassword

	// Save to database
	collection := db.MongoClient.Database("school_attendance").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, testUser)
	if err != nil {
		log.Println("DEBUG: Failed to create test user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test user"})
		return
	}

	log.Println("DEBUG: Test user created successfully")

	// Step 2: Attempt to log in with the test user
	// Create a new context to avoid cancellation from the previous one
	loginCtx, loginCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer loginCancel()

	// Try to find the user we just created
	var foundUser models.User
	err = collection.FindOne(loginCtx, bson.M{"email": testUser.Email}).Decode(&foundUser)
	if err != nil {
		log.Println("DEBUG: Failed to find test user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find test user"})
		return
	}

	log.Println("DEBUG: Found test user in database")
	log.Printf("DEBUG: Stored hash: %s", foundUser.Password)

	// Check the password
	originalPassword := "testpassword123" // The password we used to create the user
	match := utils.CheckPasswordHash(originalPassword, foundUser.Password)
	log.Println("DEBUG: Password match result:", match)

	if !match {
		log.Println("DEBUG: Password verification failed!")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password verification failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Test login successful",
		"user_id": testUser.ID,
		"email":   testUser.Email,
	})
}

func GetSchools(c *gin.Context) {
	role := c.GetString("role")
	if role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only super_admin can view schools"})
		return
	}
	collection := db.MongoClient.Database("school_attendance").Collection("schools")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schools"})
		return
	}
	defer cursor.Close(ctx)

	var schools []bson.M
	if err := cursor.All(ctx, &schools); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode schools"})
		return
	}

	c.JSON(http.StatusOK, schools)
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(c *gin.Context) {
	type PasswordUpdateInput struct {
		Email       string `json:"email"`
		NewPassword string `json:"newPassword"`
	}

	var input PasswordUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Trim inputs
	input.Email = strings.TrimSpace(input.Email)
	input.NewPassword = strings.TrimSpace(input.NewPassword)

	if input.Email == "" || input.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and new password are required"})
		return
	}

	collection := db.MongoClient.Database("school_attendance").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the user first to verify they exist
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update the password in the database
	update := bson.M{
		"$set": bson.M{
			"password": hashedPassword,
		},
	}

	_, err = collection.UpdateOne(ctx, bson.M{"email": input.Email}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
		"email":   user.Email,
	})
}

// CreateUserV2 is an alternative user creation endpoint that uses a different approach
func CreateUserV2(c *gin.Context) {
	// Get the raw JSON data
	jsonData, err := c.GetRawData()
	if err != nil {
		log.Println("DEBUG: Failed to read raw data:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	log.Println("DEBUG: Raw JSON data:", string(jsonData))

	// Define a struct to parse the JSON
	type UserInput struct {
		ID       string `json:"_id"`
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		SchoolID string `json:"school_id,omitempty"`
	}

	var input UserInput
	if err := json.Unmarshal(jsonData, &input); err != nil {
		log.Println("DEBUG: JSON unmarshal error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	log.Println("DEBUG: Parsed user data - ID:", input.ID, "Email:", input.Email, "Password length:", len(input.Password))

	// Validate required fields
	if input.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	if input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	if input.Role == "admin" && input.SchoolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "schoolId is required for admin users"})
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		log.Println("DEBUG: Password hashing error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	log.Println("DEBUG: Hashed password:", hashedPassword)

	// Create the user in the database
	user := models.User{
		ID:       input.ID,
		Name:     input.Name,
		Email:    input.Email,
		Password: hashedPassword,
		Role:     input.Role,
		SchoolID: input.SchoolID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.MongoClient.Database("school_attendance").Collection("users")
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		log.Println("DEBUG: Database insertion error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"_id":       user.ID,
		"name":      user.Name,
		"email":     user.Email,
		"role":      user.Role,
		"school_id": user.SchoolID,
	})
}

// UpdateParentSubscriptions updates the subscribedStudents for a parent user.
// Allowed by: super_admin, admin (school admin). Admin can only edit users in their own school.
func UpdateParentSubscriptions(c *gin.Context) {
	parentID := c.Param("_id")
	if parentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing user id"})
		return
	}

	// Auth context
	currentRole := c.GetString("role")
	currentSchoolID := c.GetString("school_id")

	// Only super_admin or admin can manage subscriptions
	if currentRole != "super_admin" && currentRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not allowed"})
		return
	}

	// Parse payload
	var body struct {
		StudentIDs []string `json:"studentIds"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	// Normalize / de-duplicate
	uniq := make(map[string]struct{}, len(body.StudentIDs))
	result := make([]string, 0, len(body.StudentIDs))
	for _, id := range body.StudentIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := uniq[id]; !ok {
			uniq[id] = struct{}{}
			result = append(result, id)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	usersCol := db.MongoClient.Database("school_attendance").Collection("users")

	// Load the target user
	var parent models.User
	if err := usersCol.FindOne(ctx, bson.M{"_id": parentID}).Decode(&parent); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if parent.Role != "parent" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Target user is not a parent"})
		return
	}

	// If current is admin, enforce same school
	if currentRole == "admin" {
		if parent.SchoolID == "" || parent.SchoolID != currentSchoolID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Cannot modify parent outside your school"})
			return
		}
	}

	// Optional: validate student IDs belong to the same school (if your student model has schoolId).
	// Skipping deep validation for now.

	update := bson.M{
		"$set": bson.M{
			"subscribedStudents": result,
			// Optional: you can also rebuild SubscriptionMap if you use it for notification throttling
			// "subscriptionMap": bson.M{},
		},
	}

	if _, err := usersCol.UpdateOne(ctx, bson.M{"_id": parentID}, update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update subscriptions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscriptions updated", "count": len(result)})
}
