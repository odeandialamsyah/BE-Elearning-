package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"

	"github.com/gofiber/fiber/v2"
)

func AdminOverview(c *fiber.Ctx) error {
	var totalUsers, totalCourses int64

	database.DB.Model(&models.User{}).Count(&totalUsers)
	database.DB.Model(&models.Course{}).Count(&totalCourses)

	return c.JSON(fiber.Map{
		"total_users":   totalUsers,
		"total_courses": totalCourses,
	})
}

func InstructorEarnings(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)

	var courses []models.Course
	if err := database.DB.Where("instructor_id = ?", userIDStr).Find(&courses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var report []fiber.Map

	for _, course := range courses {
		var totalEnrollments int64
		database.DB.Model(&models.Enrollment{}).
			Where("course_id = ?", course.ID).
			Count(&totalEnrollments)

		report = append(report, fiber.Map{
			"course_id":        course.ID,
			"course_title":     course.Title,
			"total_enrollment": totalEnrollments,
		})
	}

	return c.JSON(fiber.Map{
		"courses": report,
	})
}

// func InstructorCourses(c *fiber.Ctx) error {
// 	userIDStr := c.Locals("user_id").(string)

// 	var courses []models.Course
// 	if err := database.DB.Where("instructor_id = ?", userIDStr).Find(&courses).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{
// 		"courses": courses,
// 	})
// }

func InstructorCourses(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)

    var courses []models.Course

    // Ambil semua course milik instructor
    if err := database.DB.Where("instructor_id = ?", userID).
        Find(&courses).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    type CourseResponse struct {
        ID           uint   `json:"id"`
        Title        string `json:"title"`
        Description  string `json:"description"`
        Published    bool   `json:"published"`
        TotalModules int    `json:"total_modules"`
        TotalQuizzes int    `json:"total_quizzes"`
    }

    var result []CourseResponse

    for _, course := range courses {
        var moduleCount int64
        var quizCount int64

        // Hitung modul
        database.DB.Model(&models.Module{}).
            Where("course_id = ?", course.ID).
            Count(&moduleCount)

        // Hitung quiz
        database.DB.Model(&models.Quiz{}).
            Joins("JOIN modules ON modules.id = quizzes.module_id").
            Where("modules.course_id = ?", course.ID).
            Count(&quizCount)

        result = append(result, CourseResponse{
            ID:           course.ID,
            Title:        course.Title,
            Description:  course.Description,
            Published:    course.Published,
            TotalModules: int(moduleCount),
            TotalQuizzes: int(quizCount),
        })
    }

    return c.JSON(fiber.Map{
        "courses": result,
    })
}


func InstructorFeedback(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)

	var courses []models.Course
	if err := database.DB.Where("instructor_id = ?", userIDStr).Find(&courses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var courseIDs []uint
	for _, course := range courses {
		courseIDs = append(courseIDs, course.ID)
	}

	var feedbacks []fiber.Map
	if len(courseIDs) > 0 {
		var records []models.Feedback
		if err := database.DB.
			Preload("User").
			Preload("Course").
			Where("course_id IN ?", courseIDs).
			Find(&records).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		for _, fb := range records {
			feedbacks = append(feedbacks, fiber.Map{
				"id":            fb.ID,
				"user_id":       fb.UserID,
				"user_name":     fb.User.FullName,
				"course_id":     fb.CourseID,
				"course_title":  fb.Course.Title,
				"rating":        fb.Rating,
				"comment":       fb.Comment,
				"created_at":    fb.CreatedAt,
			})
		}
	}

	return c.JSON(fiber.Map{
		"feedbacks": feedbacks,
	})
}
