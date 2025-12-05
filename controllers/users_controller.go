package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func GetAllUsers(c *fiber.Ctx) error {
	var users []models.User

	if err := database.DB.Find(&users).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch users"})
	}

	return c.JSON(users)
}

func GetUserByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(user)
}

// UpdateUser -> PUT /admin/users/:id
func UpdateUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	database.DB.First(&user, id)

	return c.JSON(fiber.Map{
		"message": "user updated successfully",
		"user":    user,
	})
}

// DeleteUser -> DELETE /admin/users/:id
func DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}

	if err := database.DB.Delete(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "user deleted successfully",
	})
}

// EnrollCourse -> POST /me/courses/:id/enroll
func EnrollCourse(c *fiber.Ctx) error {
	courseIDParam := c.Params("id")
	courseID, err := strconv.ParseUint(courseIDParam, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid course ID"})
	}

	userIDStr := c.Locals("user_id").(string)
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
	}

	// Check if course exists
	var course models.Course
	if err := database.DB.First(&course, courseID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "course not found"})
	}

	// Check existing enrollment
	var existing models.Enrollment
	if err := database.DB.Where("user_id = ? AND course_id = ?", uint(userID), course.ID).First(&existing).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": "already enrolled"})
	}

	enrollment := models.Enrollment{
		UserID:   uint(userID),
		CourseID: course.ID,
	}
	if err := database.DB.Create(&enrollment).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "enrolled successfully"})
}

// GetMyCourses -> GET /me/courses
func GetMyCourses(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	var enrollments []models.Enrollment
	if err := database.DB.Preload("Course").Where("user_id = ?", uint(userID)).Find(&enrollments).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var courses []fiber.Map
	for _, e := range enrollments {
		courses = append(courses, fiber.Map{
			"id":          e.Course.ID,
			"title":       e.Course.Title,
			"description": e.Course.Description,
		})
	}

	return c.JSON(courses)
}


