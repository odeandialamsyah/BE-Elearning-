package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// CreateQuiz -> POST /courses/:course_id/modules/:module_id/quizzes
func CreateQuiz(c *fiber.Ctx) error {
	moduleIDParam := c.Params("module_id")
	moduleID, err := strconv.ParseUint(moduleIDParam, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid module ID"})
	}

	// Payload berupa array of quiz
	var payload []struct {
		Question string   `json:"question"`
		Options  []string `json:"options"`
		Answer   string   `json:"answer"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if len(payload) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no quizzes provided"})
	}

	var quizzes []models.Quiz

	for i, q := range payload {
		// Validasi minimal
		if q.Question == "" {
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("question %d is empty", i+1)})
		}
		if len(q.Options) != 4 {
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("quiz %d must have exactly 4 options", i+1)})
		}
		if q.Answer == "" {
			return c.Status(400).JSON(fiber.Map{"error": fmt.Sprintf("quiz %d missing answer", i+1)})
		}

		optionsJSON, _ := json.Marshal(q.Options)
		quiz := models.Quiz{
			ModuleID: uint(moduleID),
			Question: q.Question,
			Options:  string(optionsJSON),
			Answer:   q.Answer,
		}

		quizzes = append(quizzes, quiz)
	}

	// Simpan semua quiz dalam satu operasi
	if err := database.DB.Create(&quizzes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "quizzes created successfully",
		"data":    quizzes,
	})
}

// ListQuizzes -> GET /courses/:course_id/modules/:module_id/quizzes
func ListQuizzes(c *fiber.Ctx) error {
	moduleIDParam := c.Params("module_id")
	moduleID, err := strconv.ParseUint(moduleIDParam, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid module ID"})
	}

	var quizzes []models.Quiz
	if err := database.DB.Where("module_id = ?", moduleID).Find(&quizzes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	for i := range quizzes {
		quizzes[i].Answer = "" // hide correct answers
	}
	return c.JSON(quizzes)
}

// SubmitQuiz -> POST /courses/:course_id/modules/:id/submit
func SubmitQuiz(c *fiber.Ctx) error {
	// Ambil parameter dari URL
	courseIDParam := c.Params("course_id")
	moduleIDParam := c.Params("module_id")

	courseID, err := strconv.ParseUint(courseIDParam, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid course ID"})
	}

	moduleID, err := strconv.ParseUint(moduleIDParam, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid module ID"})
	}

	// Ambil user ID dari middleware auth
	userIDStr := c.Locals("user_id").(string)
	userID, parseErr := strconv.ParseUint(userIDStr, 10, 32)
	if parseErr != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
	}

	// Payload berisi array JSON
	var payload []struct {
		QuizID uint   `json:"quiz_id"`
		Answer string `json:"answer"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	if len(payload) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "no quiz answers provided"})
	}

	// Ambil quiz dari database berdasarkan module dan ID
	quizIDs := make([]uint, len(payload))
	for i, answer := range payload {
		quizIDs[i] = answer.QuizID
	}

	var quizzes []models.Quiz
	if err := database.DB.Where("id IN ? AND module_id = ?", quizIDs, moduleID).Find(&quizzes).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(quizzes) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "no quizzes found for this module"})
	}

	// Map untuk jawaban benar
	quizAnswerMap := make(map[uint]string)
	for _, quiz := range quizzes {
		quizAnswerMap[quiz.ID] = quiz.Answer
	}

	// Hitung skor
	totalCorrect := 0
	for _, answer := range payload {
		if correctAnswer, exists := quizAnswerMap[answer.QuizID]; exists && correctAnswer == answer.Answer {
			totalCorrect++
		}
	}

	// Hitung skor dalam persen
	totalQuestions := len(quizzes)
	scorePercent := (float64(totalCorrect) / float64(totalQuestions)) * 100
	passed := scorePercent >= 60.0 // lulus jika >= 60%

	// Simpan hasil
	result := models.QuizResult{
		UserID:   uint(userID),
		ModuleID: uint(moduleID),
		Score:    int(scorePercent),
		Passed:   passed,
	}

	if err := database.DB.Create(&result).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"message":     "quiz submitted successfully",
		"course_id":   courseID,
		"module_id":   moduleID,
		"score":       result.Score,
		"passed":      result.Passed,
		"total_quiz":  totalQuestions,
		"correct":     totalCorrect,
		"wrong":       totalQuestions - totalCorrect,
	})
}

// GetQuizResults -> GET /me/courses/:course_id/modules/quiz-results
func GetQuizResults(c *fiber.Ctx) error {
	// Convert user_id from string to uint
	userIDStr := c.Locals("user_id").(string)
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
	}

	// Adjust query to fetch results for all modules in the course
	courseIDParam := c.Params("id")
	courseID, err := strconv.ParseUint(courseIDParam, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid course ID"})
	}

	var results []models.QuizResult
	if err := database.DB.Joins("JOIN modules ON modules.id = quiz_results.module_id").Where("modules.course_id = ? AND quiz_results.user_id = ?", courseID, userID).Find(&results).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(results)
}

// GetCourseStatus -> GET /me/courses/:id/status
func GetCourseStatus(c *fiber.Ctx) error {
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

	// Ambil semua modul dalam course
	var modules []models.Module
	if err := database.DB.Where("course_id = ?", courseID).Find(&modules).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	if len(modules) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "no modules found in this course"})
	}

	totalModules := len(modules)
	passedModules := 0
	var detailedResults []fiber.Map

	for _, m := range modules {
		var result models.QuizResult
		err := database.DB.
			Where("user_id = ? AND module_id = ?", userID, m.ID).
			Order("created_at desc").
			First(&result).Error

		moduleStatus := "Not Attempted"
		score := 0

		if err == nil {
			score = result.Score
			if result.Passed {
				passedModules++
				moduleStatus = "Passed"
			} else {
				moduleStatus = "Failed"
			}
		}

		detailedResults = append(detailedResults, fiber.Map{
			"module_id":    m.ID,
			"module_title": m.Title,
			"score":        score,
			"status":       moduleStatus,
		})
	}

	completed := passedModules == totalModules

	return c.JSON(fiber.Map{
		"total_modules":  totalModules,
		"passed_modules": passedModules,
		"completion_rate": fmt.Sprintf("%.2f%%", float64(passedModules)/float64(totalModules)*100),
		"completed":      completed,
		"details":        detailedResults,
	})
}