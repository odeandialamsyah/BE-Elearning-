package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"

	"github.com/gofiber/fiber/v2"
)

// AddModuleToCourse -> POST /instructor/courses/:id/modules (instructor only)
func AddModuleToCourse(c *fiber.Ctx) error {
	courseID := c.Params("id")

	var payload struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Order   int    `json:"order"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// ensure course exists
	var course models.Course
	if err := database.DB.First(&course, courseID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "course not found"})
	}

	module := models.Module{
		Title:   payload.Title,
		Content: payload.Content,
		Order:   payload.Order,
		CourseID: course.ID,
	}

	if err := database.DB.Create(&module).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(module)
}