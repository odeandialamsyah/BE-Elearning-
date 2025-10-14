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

// EditModule -> PUT /instructor/courses/:course_id/modules/:id (requires instructor)
func EditModule(c *fiber.Ctx) error {
	id := c.Params("id")

	var payload struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Order   int    `json:"order"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var module models.Module
	if err := database.DB.First(&module, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "module not found"})
	}

	// Update fields
	module.Title = payload.Title
	module.Content = payload.Content
	module.Order = payload.Order

	if err := database.DB.Save(&module).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(module)
}

// DeleteModule -> DELETE /instructor/courses/:course_id/modules/:id (requires instructor)
func DeleteModule(c *fiber.Ctx) error {
	id := c.Params("id")

	var module models.Module
	if err := database.DB.First(&module, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "module not found"})
	}

	if err := database.DB.Delete(&module).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(204).Send(nil)
}