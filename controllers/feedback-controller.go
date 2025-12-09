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

	// Ambil user_id sebagai string lalu convert ke uint
	uidStr := c.Locals("user_id").(string)

	uidUint, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid user ID"})
	}

	userID := uint(uidUint)

	feedback := models.Feedback{
		UserID:   userID,
		CourseID: input.CourseID,
		Rating:   input.Rating,
		Comment:  input.Comment,
	}

	if err := database.DB.Create(&feedback).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan feedback"})
	}

	// Reload feedback with associations so User and Course are populated in response
	if err := database.DB.Preload("User").Preload("Course").First(&feedback, feedback.ID).Error; err != nil {
		// If preload fails, still return created feedback (without relations)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Feedback berhasil dikirim",
			"data":    feedback,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Feedback berhasil dikirim",
		"data":    feedback,
	})
}

func GetAllFeedback(c *fiber.Ctx) error {
	// Query parameters
	courseID := c.Query("course_id")
	instructorID := c.Query("instructor_id")
	userID := c.Query("user_id")
	rating := c.Query("rating")
	minRating := c.Query("min_rating")
	maxRating := c.Query("max_rating")
	sort := c.Query("sort", "desc") // default newest first

	var feedbacks []models.Feedback
	query := database.DB.Preload("User").Preload("Course")

	// Filter by course
	if courseID != "" {
		query = query.Where("course_id = ?", courseID)
	}

	// Filter by instructor (join ke tabel course)
	if instructorID != "" {
		query = query.Joins("JOIN courses ON courses.id = feedback.course_id").
			Where("courses.instructor_id = ?", instructorID)
	}

	// Filter by user
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Rating exact
	if rating != "" {
		query = query.Where("rating = ?", rating)
	}

	// Rating minimum
	if minRating != "" {
		query = query.Where("rating >= ?", minRating)
	}

	// Rating maksimum
	if maxRating != "" {
		query = query.Where("rating <= ?", maxRating)
	}

	// Sorting
	if sort == "asc" {
		query = query.Order("created_at ASC")
	} else {
		query = query.Order("created_at DESC")
	}

	// Execute
	if err := query.Find(&feedbacks).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch feedback"})
	}

	// Build response with course name, instructor name and date
	var out []fiber.Map
	for _, fb := range feedbacks {
		courseTitle := ""
		instructorName := ""
		if fb.Course.ID != 0 {
			courseTitle = fb.Course.Title
			// load instructor
			var instr models.User
			if fb.Course.InstructorID != 0 {
				if err := database.DB.First(&instr, fb.Course.InstructorID).Error; err == nil {
					instructorName = instr.FullName
				}
			}
		}

		userName := ""
		if fb.User.ID != 0 {
			userName = fb.User.FullName
		}

		out = append(out, fiber.Map{
			"id":              fb.ID,
			"user_id":         fb.UserID,
			"user_name":       userName,
			"course_id":       fb.CourseID,
			"course_title":    courseTitle,
			"instructor_name": instructorName,
			"rating":          fb.Rating,
			"comment":         fb.Comment,
			"created_at":      fb.CreatedAt,
		})
	}

	return c.JSON(fiber.Map{
		"total":     len(out),
		"feedbacks": out,
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