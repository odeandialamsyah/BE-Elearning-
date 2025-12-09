package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
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

// setting user profile
func UpdateProfile(c *fiber.Ctx) error {
    uid := c.Locals("user_id")
    if uid == nil {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }

    // 🔥 convert string → uint
    uidStr := uid.(string)
    userID, err := strconv.ParseUint(uidStr, 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID"})
    }

    var body struct {
        FullName string `json:"full_name"`
        Email    string `json:"email"`
        Username string `json:"username"`
    }

    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
    }

    var user models.User
    if err := database.DB.First(&user, uint(userID)).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "user not found"})
    }

    user.FullName = body.FullName
    user.Email = body.Email
    user.Username = body.Username

    database.DB.Save(&user)

    return c.JSON(fiber.Map{
        "message": "profile updated",
        "user":    user,
    })
}

func ChangePassword(c *fiber.Ctx) error {
    uid := c.Locals("user_id")
    if uid == nil {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }

    uidStr := uid.(string)
    userID, err := strconv.ParseUint(uidStr, 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID"})
    }

    var body struct {
        OldPassword string `json:"old_password"`
        NewPassword string `json:"new_password"`
    }

    if err := c.BodyParser(&body); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
    }

    var user models.User
    if err := database.DB.First(&user, uint(userID)).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "user not found"})
    }

    // cek password lama
    if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.OldPassword)) != nil {
        return c.Status(400).JSON(fiber.Map{"error": "old password is incorrect"})
    }

    // hash password baru
    hashed, _ := bcrypt.GenerateFromPassword([]byte(body.NewPassword), 14)
    user.Password = string(hashed)

    database.DB.Save(&user)

    return c.JSON(fiber.Map{"message": "password updated"})
}

func DeleteAccount(c *fiber.Ctx) error {
    uid := c.Locals("user_id")
    if uid == nil {
        return c.Status(401).JSON(fiber.Map{"error": "unauthorized"})
    }

    uidStr := uid.(string)
    userID, err := strconv.ParseUint(uidStr, 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid User ID"})
    }

    if err := database.DB.Delete(&models.User{}, uint(userID)).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "failed to delete account"})
    }

    return c.JSON(fiber.Map{"message": "account deleted"})
}
