package main

import (
	"school-attendance/controllers"
	"school-attendance/db" // add this
	"school-attendance/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db.ConnectMongo()
	r := gin.Default()
	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Public routes
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "School Attendance API is running 🚀"})
	})
	r.POST("/login", controllers.Login)
	r.POST("/users", controllers.CreateUser)
	r.POST("/users/v2", controllers.CreateUserV2)
	r.GET("/test-login", controllers.TestLoginEndpoint)
	r.POST("/update-password", controllers.UpdateUserPassword)

	r.GET("/terms/:id", controllers.GetTerm)
	r.GET("/classrooms/:id", controllers.GetClassroom)
	r.GET("/students/classroom/:classroom_id/term/:term_id", controllers.GetStudentsByClassroomAndTerm)
	r.POST("/classrooms/:id/sessions", controllers.AddSessionToClassroom)
	r.GET("/classrooms/:id/sessions", controllers.GetSessionsForClassroom)
	r.DELETE("/classrooms/:id/sessions/:session_id", controllers.RemoveSessionFromClassroom)
	r.POST("/attendance/check", controllers.CheckInOrOutStudent)
	r.GET("/students/classroom/:classroom_id", controllers.GetStudentsByClassroom)
	// r.GET("/schools/:id", controllers.GetSchool) is use to fetch only school infor for teacher
	r.GET("/schools/:id", controllers.GetSchool)
	r.GET("/attendance/session/:session_id", controllers.GetAttendanceBySession)
	// Protected routes
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
		protected.GET("/grades", controllers.GetGrades)
		protected.PUT("/grades/:id", controllers.UpdateGrade)
		protected.POST("/classrooms", controllers.CreateClassroom)
		protected.GET("/classrooms", controllers.GetClassrooms)
		protected.PUT("/classrooms/:id", controllers.UpdateClassroom)
		protected.PUT("/classrooms/:id/assign-teacher", controllers.AssignTeacherToClassroom)
		protected.PUT("/classrooms/:id/assign-students", controllers.AssignStudentsToClassroom)
		protected.PUT("/students/:id/classroom", controllers.UpdateStudentClassroom)
		protected.POST("/students", controllers.CreateStudent)
		protected.GET("/students", controllers.GetStudents)
		protected.PUT("/students/:id", controllers.UpdateStudent)
		protected.GET("/parents/me/children", controllers.GetMyChildren)
		protected.GET("/parents/me/children/:student_id/attendance", controllers.GetMyChildAttendance)
		// protected.GET("/parents/:parentId/children", controllers.GetChildrenForParent)

		protected.GET("/attendance/student/:student_id", controllers.GetAttendanceByStudent)

		protected.PUT("/users/:id/subscriptions", controllers.UpdateParentSubscriptions)
		protected.PUT("/users/:id", controllers.AdminUpdateUser)
		protected.PUT("/users/:id/password", controllers.AdminSetUserPassword)
		// Teacher report endpoints
		protected.GET("/attendance/by-classroom/:classroom_id", controllers.AttendanceByClassroom) // ?date=YYYY-MM-DD or ?from=&to=
		// protected.GET("/attendance/student/:studentId", controllers.AttendanceByStudent)          // ?from=&to=
		protected.GET("/attendance/range", controllers.AttendanceRangeForTeacher) // ?from=&to=
		protected.GET("/attendance/today", controllers.AttendanceTodayForTeacher) // today only
	}

	r.Run(":8080")
}
