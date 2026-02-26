package routes

import (
	"arlchoose/backend-api/controllers"
	"arlchoose/backend-api/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {

	// Initialize gin
	router := gin.Default()

	// Setup CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
	}))

	// Base API group
	api := router.Group("/api")

	// Auth routes
	api.POST("/login", controllers.Login)

	// Authenticated routes
	auth := api.Group("/")
	auth.Use(middlewares.AuthMiddleware())
	{
		// Upload
		auth.POST("/upload", controllers.UploadFile)
		auth.DELETE("/upload", controllers.DeleteFile)

		auth.POST("/register", controllers.Register)

		// Users
		auth.GET("/users", controllers.FindUsers)
		auth.POST("/users", controllers.CreateUser)
		auth.GET("/users/:id", controllers.FindUserById)
		auth.PUT("/users/:id", controllers.UpdateUser)
		auth.DELETE("/users/:id", controllers.DeleteUser)

		// Contacts — admin
		auth.GET("/contacts", controllers.FindContacts)
		auth.GET("/contacts/:id", controllers.FindContactById)
		auth.PUT("/contacts/:id/status", controllers.UpdateContactStatus)
		auth.DELETE("/contacts/:id", controllers.DeleteContact)

		// Skills — auth
		auth.POST("/skills", controllers.CreateSkill)
		auth.PUT("/skills/:id", controllers.UpdateSkill)
		auth.DELETE("/skills/:id", controllers.DeleteSkill)

		// Educations — auth
		auth.POST("/educations", controllers.CreateEducation)
		auth.PUT("/educations/:id", controllers.UpdateEducation)
		auth.DELETE("/educations/:id", controllers.DeleteEducation)

		// Courses — auth
		auth.POST("/courses", controllers.CreateCourse)
		auth.PUT("/courses/:id", controllers.UpdateCourse)
		auth.DELETE("/courses/:id", controllers.DeleteCourse)

		// Experiences — auth
		auth.POST("/experiences", controllers.CreateExperience)
		auth.PUT("/experiences/:id", controllers.UpdateExperience)
		auth.DELETE("/experiences/:id", controllers.DeleteExperience)
		auth.POST("/experiences/:id/images", controllers.AddExperienceImage)
		auth.DELETE("/experiences/:id/images/:imageId", controllers.DeleteExperienceImage)

		// Projects -auth
		auth.POST("/projects", controllers.CreateProject)
		auth.PUT("/projects/:id", controllers.UpdateProject)
		auth.DELETE("/projects/:id", controllers.DeleteProject)
		auth.POST("/projects/:id/images", controllers.AddProjectImage)
		auth.DELETE("/projects/:id/images/:imageId", controllers.DeleteProjectImage)

		auth.POST("/tags", controllers.CreateTag)
		auth.PUT("/tags/:id", controllers.UpdateTag)
		auth.DELETE("/tags/:id", controllers.DeleteTag)

		auth.POST("/blogs", controllers.CreateBlog)
		auth.GET("/blogs/all", controllers.FindAllBlogs)
		auth.PUT("/blogs/:id", controllers.UpdateBlog)
		auth.DELETE("/blogs/:id", controllers.DeleteBlog)

		// AI Blog generation
		auth.POST("/blogs/generate", controllers.GenerateAiBlog)
		auth.PUT("/blogs/:id/publish", controllers.PublishBlog)
		auth.PUT("/blogs/:id/reject", controllers.RejectBlog)

		auth.POST("/bookmarks/sync", controllers.SyncAllBookmarks)
		auth.DELETE("/bookmarks/:id", controllers.DeleteBookmark)

		auth.PUT("/profile", controllers.UpsertProfile)
		auth.PUT("/settings", controllers.UpsertSettings)

		auth.GET("/tools/all", controllers.FindAllTools)
		auth.POST("/tools", controllers.CreateTool)
		auth.PUT("/tools/:id", controllers.UpdateTool)
		auth.PUT("/tools/:id/toggle", controllers.ToggleTool)
		auth.DELETE("/tools/:id", controllers.DeleteTool)

		auth.GET("/blogs/stream", controllers.BlogStream)

		auth.GET("/blogs/stats", controllers.BlogStats)
		auth.PUT("/blogs/:id/archive", controllers.ArchiveBlog)
		auth.POST("/blogs/bulk", controllers.BulkActionBlog)

	}

	// Public routes
	public := api.Group("/")
	{
		// Contacts — publik
		public.POST("/contacts", middlewares.ContactRateLimit(), controllers.CreateContact)

		// Skills — publik
		public.GET("/skills", controllers.FindSkills)
		public.GET("/skills/:id", controllers.FindSkillById)

		// Educations — publik
		public.GET("/educations", controllers.FindEducations)
		public.GET("/educations/:id", controllers.FindEducationById)

		// Courses — publik
		public.GET("/courses", controllers.FindCourses)
		public.GET("/courses/:id", controllers.FindCourseById)

		// Experiences - publik
		public.GET("/experiences", controllers.FindExperiences)
		public.GET("/experiences/:id", controllers.FindExperienceById)

		// Projects - publik
		public.GET("/projects", controllers.FindProjects)
		public.GET("/projects/:slug", controllers.FindProjectBySlug)

		public.GET("/tags", controllers.FindTags)
		public.GET("/tags/:id", controllers.FindTagById)
		public.GET("/tags/slug/:slug", controllers.FindTagBySlug)

		public.GET("/blogs", controllers.FindBlogs)
		public.GET("/blogs/:slug", controllers.FindBlogBySlug)

		public.GET("/bookmarks", controllers.FindBookmarks)
		public.GET("/bookmarks/:id", controllers.FindBookmarkById)

		public.GET("/profile", controllers.GetProfile)
		public.GET("/settings", controllers.GetSettings)
		public.GET("/settings/:key", controllers.GetSettingByKey)

		public.GET("/tools", middlewares.ToolRateLimit(), controllers.FindTools)
		public.GET("/tools/registry", controllers.FindRegistry) // ← HARUS DI ATAS :slug
		public.GET("/tools/:slug", middlewares.ToolRateLimit(), controllers.FindToolBySlug)
		public.GET("/tools/:slug/run", middlewares.ToolRateLimit(), controllers.RunTool)
		public.POST("/tools/:slug/run", middlewares.ToolRateLimit(), controllers.RunTool)
	}

	return router
}
