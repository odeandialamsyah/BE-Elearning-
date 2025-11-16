package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"

	"github.com/gofiber/fiber/v2"
)

func AdminOverview(c *fiber.Ctx) error {
	var totalUsers, totalCourses, totalOrders int64
	var totalRevenue float64

	database.DB.Model(&models.User{}).Count(&totalUsers)
	database.DB.Model(&models.Course{}).Count(&totalCourses)
	database.DB.Model(&models.Order{}).Count(&totalOrders)
	database.DB.Model(&models.Order{}).Where("status = ?", "paid").Select("SUM(amount)").Scan(&totalRevenue)

	return c.JSON(fiber.Map{
		"total_users":     totalUsers,
		"total_courses":   totalCourses,
		"total_orders":    totalOrders,
		"total_revenue":   totalRevenue,
	})
}

func AdminTransactions(c *fiber.Ctx) error {
	var orders []models.Order
	if err := database.DB.Preload("User").Preload("Course").Find(&orders).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(orders)
}

func InstructorEarnings(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)

	var courses []models.Course
	if err := database.DB.Where("instructor_id = ?", userIDStr).Find(&courses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var totalEarnings float64
	var report []fiber.Map

	for _, course := range courses {
		var courseRevenue float64
		var totalEnrollments int64
		database.DB.Model(&models.Order{}).
			Where("course_id = ? AND status = ?", course.ID, "paid").
			Select("SUM(amount)").Scan(&courseRevenue)
		database.DB.Model(&models.Enrollment{}).
			Where("course_id = ?", course.ID).
			Count(&totalEnrollments)

		totalEarnings += courseRevenue
		report = append(report, fiber.Map{
			"course_id":       course.ID,
			"course_title":    course.Title,
			"revenue":         courseRevenue,
			"total_enrollment": totalEnrollments,
		})
	}

	return c.JSON(fiber.Map{
		"total_earnings": totalEarnings,
		"courses":        report,
	})
}

func InstructorCourses(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)

	var courses []models.Course
	if err := database.DB.Where("instructor_id = ?", userIDStr).Find(&courses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"courses": courses,
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
