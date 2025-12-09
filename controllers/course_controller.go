package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// CreateCourse -> POST /instructor/courses (requires instructor)
func CreateCourse(c *fiber.Ctx) error {
	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	instructorID := c.Locals("user_id")

	course := models.Course{
		Title:       payload.Title,
		Description: payload.Description,
	}
	// middleware stores user_id as string; attempt to parse
	if sid, ok := instructorID.(string); ok {
		if uid, err := strconv.ParseUint(sid, 10, 64); err == nil {
			course.InstructorID = uint(uid)
		}
	}

	if err := database.DB.Create(&course).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(course)
}

// ListPublishedCourses -> GET /courses
func ListPublishedCourses(c *fiber.Ctx) error {
	var courses []models.Course
	if err := database.DB.Where("published = ?", true).Find(&courses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var out []fiber.Map
	for _, course := range courses {
		var instructor models.User
		instructorName := ""
		if course.InstructorID != 0 {
			if err := database.DB.First(&instructor, course.InstructorID).Error; err == nil {
				instructorName = instructor.FullName
			}
		}
		out = append(out, fiber.Map{
			"id":          course.ID,
			"title":       course.Title,
			"description": course.Description,
			"published":   course.Published,
			"instructor": fiber.Map{"id": course.InstructorID, "name": instructorName},
		})
	}

	return c.JSON(out)
}

// ListUnpublishedCourses -> GET /admin/courses/unpublished (admin only)
func ListUnpublishedCourses(c *fiber.Ctx) error {
	var courses []models.Course
	if err := database.DB.Where("published = ?", false).Find(&courses).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	var out []fiber.Map
	for _, course := range courses {
		var instructor models.User
		instructorName := ""
		if course.InstructorID != 0 {
			if err := database.DB.First(&instructor, course.InstructorID).Error; err == nil {
				instructorName = instructor.FullName
			}
		}
		out = append(out, fiber.Map{
			"id":          course.ID,
			"title":       course.Title,
			"description": course.Description,
			"published":   course.Published,
			"instructor": fiber.Map{"id": course.InstructorID, "name": instructorName},
		})
	}

	return c.JSON(out)
}

// ListAllCoursesByStatus -> GET /admin/courses/status (admin only)
// returns grouped courses by published status
func ListAllCoursesByStatus(c *fiber.Ctx) error {
	var published []models.Course
	var unpublished []models.Course

	if err := database.DB.Where("published = ?", true).Find(&published).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if err := database.DB.Where("published = ?", false).Find(&unpublished).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	build := func(courses []models.Course) []fiber.Map {
		var res []fiber.Map
		for _, course := range courses {
			var instructor models.User
			instructorName := ""
			if course.InstructorID != 0 {
				if err := database.DB.First(&instructor, course.InstructorID).Error; err == nil {
					instructorName = instructor.FullName
				}
			}
			res = append(res, fiber.Map{
				"id":          course.ID,
				"title":       course.Title,
				"description": course.Description,
				"published":   course.Published,
				"instructor":  fiber.Map{"id": course.InstructorID, "name": instructorName},
			})
		}
		return res
	}

	return c.JSON(fiber.Map{
		"published":   build(published),
		"unpublished": build(unpublished),
	})
}

func GetCourseDetail(c *fiber.Ctx) error {
    courseID := c.Params("id")

    uid := c.Locals("user_id")
    if uid == nil {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }
    userID := uid.(string)

    // Preload modules
    var course models.Course
    err := database.DB.
        Preload("Modules", func(db *gorm.DB) *gorm.DB {
            return db.Order("`order` ASC")
        }).
        First(&course, courseID).Error

    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "course not found"})
    }

    // Cek owner
    if strconv.Itoa(int(course.InstructorID)) != userID {
        return c.Status(403).JSON(fiber.Map{"error": "not your course"})
    }

    // Ambil quiz per modul
    type ModuleResponse struct {
        ID      uint          `json:"id"`
        Title   string        `json:"title"`
        PDFUrl  string        `json:"pdf_url"`
        Order   int           `json:"order"`
        Quizzes []models.Quiz `json:"quizzes"`
    }

    var modulesWithQuiz []ModuleResponse

    for _, m := range course.Modules {
        var quizzes []models.Quiz
        database.DB.Where("module_id = ?", m.ID).Find(&quizzes)

        modulesWithQuiz = append(modulesWithQuiz, ModuleResponse{
            ID:      m.ID,
            Title:   m.Title,
            PDFUrl:  m.PDFUrl,
            Order:   m.Order,
            Quizzes: quizzes,
        })
    }

    return c.JSON(fiber.Map{
        "course": fiber.Map{
            "id":          course.ID,
            "title":       course.Title,
            "description": course.Description,
            "published":   course.Published,
        },
        "modules": modulesWithQuiz,
    })
}


// PublishCourse -> PUT /instructor/courses/:id/publish (admin only)
func PublishCourse(c *fiber.Ctx) error {
	id := c.Params("id")
	var course models.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "course not found"})
	}
	course.Published = true
	if err := database.DB.Save(&course).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(course)
}

// UnpublishCourse -> PUT /instructor/courses/:id/unpublish (admin only)
func UnpublishCourse(c *fiber.Ctx) error {
	id := c.Params("id")
	var course models.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "course not found"})
	}
	course.Published = false
	if err := database.DB.Save(&course).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(course)
}

// EditCourse -> PUT /instructor/courses/:id (requires instructor)
func EditCourse(c *fiber.Ctx) error {
	id := c.Params("id")
	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var course models.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "course not found"})
	}

	// Update fields
	course.Title = payload.Title
	course.Description = payload.Description

	if err := database.DB.Save(&course).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(course)
}

// DeleteCourse -> DELETE /instructor/courses/:id (requires instructor)
func DeleteCourse(c *fiber.Ctx) error {
	id := c.Params("id")

	var course models.Course
	if err := database.DB.First(&course, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "course not found"})
	}

	if err := database.DB.Delete(&course).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(204).Send(nil)
}