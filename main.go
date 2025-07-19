package main

import (
	"school-attendance/controllers"
	"school-attendance/db" // add this
	"school-attendance/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	db.ConnectMongo() // connect to MongoDB
	r := gin.Default()

	// ✅ Enable CORS for frontend requests
	// r.Use(cors.Default())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "School Attendance API is running 🚀"})
	})
	r.POST("/login", controllers.Login)
	r.POST("/users", controllers.CreateUser)
	r.POST("/users/v2", controllers.CreateUserV2)
	r.GET("/test-login", controllers.TestLoginEndpoint)
	r.POST("/update-password", controllers.UpdateUserPassword)
	r.GET("/grades", middleware.AuthMiddleware(), controllers.GetGrades)
	r.PUT("/grades/:id", middleware.AuthMiddleware(), controllers.UpdateGrade)
	r.PUT("/classrooms/:id/assign-students", controllers.AssignStudentsToClassroom)

	// 🔐 Protected Routes
	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/users", controllers.GetAllUsers)
		protected.POST("/schools", controllers.CreateSchool)
		protected.GET("/schools", controllers.GetSchools)
		protected.POST("/departments", controllers.CreateDepartment)
		protected.GET("/departments", controllers.GetDepartments)
		protected.PUT("/departments/:id", controllers.UpdateDepartment)
		protected.POST("/terms", controllers.CreateTerm)
		protected.GET("/terms", controllers.GetTerms)
		protected.POST("/grades", controllers.CreateGrade)
		protected.POST("/classrooms", controllers.CreateClassroom)
		protected.GET("/classrooms", controllers.GetClassrooms)
		protected.PUT("/classrooms/:id", controllers.UpdateClassroom)
		protected.PUT("/students/:id/classroom", controllers.UpdateStudentClassroom)
		protected.POST("/attendance", controllers.RecordAttendance)
		protected.POST("/students", controllers.CreateStudent)
		protected.GET("/students", controllers.GetStudents)
		protected.PUT("/students/:id", controllers.UpdateStudent)

		// more protected routes later...
	}
	// r.GET("/students", controllers.GetStudents)
	// For terms
	r.GET("/terms/:id", controllers.GetTerm)
	// r.PUT("/students/:id/classroom", controllers.UpdateStudentClassroom)
	// For classrooms
	r.GET("/classrooms/:id", controllers.GetClassroom)
	// get students by classroom id and term id , at case when want to show classroom detail who student has been study at classroom (display at frontend classroom detail page)
	r.GET("/students/classroom/:classroomId/term/:termId", controllers.GetStudentsByClassroomAndTerm)
	// r.Use(cors.Default())
	r.Group("/")
	r.Run(":8080") // listen on :8080

}
