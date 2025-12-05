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
	return c.JSON(courses)
}

func GetCourseDetail(c *fiber.Ctx) error {
    courseID := c.Params("id")

    // Ambil user_id dari JWT (AMAN TANPA PANIC)
    uid := c.Locals("user_id")
    if uid == nil {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }
    userID := uid.(string)

    // Ambil course + preload modules
    var course models.Course
    err := database.DB.
        Preload("Modules", func(db *gorm.DB) *gorm.DB {
            return db.Order("`order` ASC")
        }).
        First(&course, courseID).Error

    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "course not found"})
    }

    // Pastikan instructor benar
    if strconv.Itoa(int(course.InstructorID)) != userID {
        return c.Status(403).JSON(fiber.Map{"error": "not your course"})
    }

    // Response untuk FE
    type ModuleResponse struct {
        ID     uint   `json:"id"`
        Title  string `json:"title"`
        PDFUrl string `json:"pdf_url"`
        Order  int    `json:"order"`
    }

    var moduleList []ModuleResponse
    for _, m := range course.Modules {
        moduleList = append(moduleList, ModuleResponse{
            ID:     m.ID,
            Title:  m.Title,
            PDFUrl: m.PDFUrl,
			Order:  m.Order,
        })
    }

    return c.JSON(fiber.Map{
        "course": fiber.Map{
            "id":          course.ID,
            "title":       course.Title,
            "description": course.Description,
            "published":   course.Published,
        },
        "modules": moduleList,
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