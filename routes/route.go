package routes

import (
	"backend-elearning/controllers"
	"backend-elearning/middlewares"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/auth/register", controllers.Register)
	api.Post("/auth/login", controllers.Login)
	api.Post("/auth/logout", controllers.Logout)
	api.Get("/auth/me", middlewares.AuthMiddleware(), controllers.Profile)
	api.Post("/feedback", middlewares.AuthMiddleware(), controllers.SubmitFeedback)

	// public course listing
	api.Get("/courses", controllers.ListPublishedCourses)

	// public: list quizzes for a module (answers hidden)
	api.Get("/courses/:course_id/modules/:module_id/quizzes", controllers.ListQuizzes)

	// instructor routes (require auth + instructor role)
	instr := api.Group("/instructor", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("instructor"))
	instr.Get("/courses/:id", controllers.GetCourseDetail)
	instr.Post("/courses", controllers.CreateCourse)
	instr.Post("/courses/:course_id/modules", controllers.AddModuleToCourse)
	instr.Put("/courses/:id", controllers.EditCourse)
	instr.Delete("/courses/:id", controllers.DeleteCourse)
	instr.Put("/courses/:course_id/modules/:module_id", controllers.EditModule)
	instr.Delete("/courses/:course_id/modules/:module_id", controllers.DeleteModule)
	// quiz routes
	quiz := instr.Group("/courses/:course_id/modules/:module_id")
	quiz.Post("/quizzes", controllers.CreateQuiz)
	quiz.Get("/quizzes", controllers.ListQuizzes)
	quiz.Post("/submit", controllers.SubmitQuiz)

	instr.Get("/courses", controllers.InstructorCourses)
	instr.Get("/earnings", controllers.InstructorEarnings)
	instr.Get("/feedback", controllers.InstructorFeedback)
	// admin publish/unpublish (require auth + admin role)
	admin := api.Group("/admin", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	admin.Put("/courses/:id/publish", controllers.PublishCourse)
	admin.Put("/courses/:id/unpublish", controllers.UnpublishCourse)
	admin.Get("/overview", controllers.AdminOverview)
	admin.Get("/courses/:course_id/feedback", controllers.GetFeedbackByCourse)
	admin.Get("/feedback", controllers.GetAllFeedback)

	admin.Get("/users", controllers.GetAllUsers)
	admin.Get("/users/:id", controllers.GetUserByID)
	admin.Put("/users/:id", controllers.UpdateUser)
	admin.Delete("/users/:id", controllers.DeleteUser)

	me := api.Group("/me", middlewares.AuthMiddleware())
	me.Get("/courses/:id/modules/quiz-results", controllers.GetQuizResults)
	me.Get("/courses/:id/status", controllers.GetCourseStatus)
	me.Get("/courses", controllers.GetMyCourses)
	me.Post("/courses/:id/enroll", controllers.EnrollCourse)
}
