package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// Submit feedback untuk course
func SubmitFeedback(c *fiber.Ctx) error {
	var input struct {
		CourseID uint   `json:"course_id"`
		Rating   int    `json:"rating"`
		Comment  string `json:"comment"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Ambil user ID dari context (middleware auth)
	userID := c.Locals("user_id").(uint)

	feedback := models.Feedback{
		UserID:   userID,
		CourseID: input.CourseID,
		Rating:   input.Rating,
		Comment:  input.Comment,
	}

	if err := database.DB.Create(&feedback).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan feedback"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Feedback berhasil dikirim",
		"data":    feedback,
	})
}

// Ambil semua feedback untuk course tertentu
func GetFeedbackByCourse(c *fiber.Ctx) error {
	courseID, err := strconv.Atoi(c.Params("course_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ID course tidak valid"})
	}

	var feedbacks []models.Feedback
	if err := database.DB.Preload("User").Where("course_id = ?", courseID).Find(&feedbacks).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal mengambil data feedback"})
	}

	return c.JSON(fiber.Map{
		"course_id": courseID,
		"feedbacks": feedbacks,
	})
}