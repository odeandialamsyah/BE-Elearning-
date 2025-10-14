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
	api.Get("/auth/me", middlewares.AuthMiddleware(), controllers.Profile)

	// public course listing
	api.Get("/courses", controllers.ListPublishedCourses)
	api.Get("/courses/:id", controllers.GetCourseDetail)

	// instructor routes (require auth + instructor role)
	instr := api.Group("/instructor", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("instructor"))
	instr.Post("/courses", controllers.CreateCourse)
	instr.Post("/courses/:id/modules", controllers.AddModuleToCourse)
	instr.Put("/courses/:id", controllers.EditCourse)
	instr.Delete("/courses/:id", controllers.DeleteCourse)
	instr.Put("/courses/:id/modules/:id", controllers.EditModule)
	instr.Delete("/courses/:id/modules/:id", controllers.DeleteModule)

	// admin publish/unpublish (require auth + admin role)
	admin := api.Group("/admin", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
	admin.Put("/courses/:id/publish", controllers.PublishCourse)
	admin.Put("/courses/:id/unpublish", controllers.UnpublishCourse)

	// quiz routes
	quiz := instr.Group("/courses/:course_id/modules/:module_id")
	quiz.Post("/quizzes", controllers.CreateQuiz)
	quiz.Get("/quizzes", controllers.ListQuizzes)
	quiz.Post("/submit", controllers.SubmitQuiz)

	me := api.Group("/me", middlewares.AuthMiddleware())
	me.Get("/courses/:id/modules/quiz-results", controllers.GetQuizResults)
	me.Get("/courses/:id/status", controllers.GetCourseStatus)
}
