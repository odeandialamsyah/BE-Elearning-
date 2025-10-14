package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"
	"backend-elearning/utils"

	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// Register
func Register(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	// hash password
	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 14)
	user.Password = string(hash)

	// default role
	if user.Role == "" {
		user.Role = "user"
	}

	if err := database.DB.Create(&user).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "username/email exists"})
	}

	return c.JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login
func Login(c *fiber.Ctx) error {
	data := new(struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	})
	if err := c.BodyParser(data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var user models.User
	if err := database.DB.Where("email = ?", data.Email).First(&user).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(data.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := utils.GeneratePasetoToken(user.ID, user.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "token generation error", "details": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "login successful",
		"token":   token,
		"user": fiber.Map{
			"id":       user.ID,
			"fullName": user.FullName,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// Profile (butuh middleware auth)
func Profile(c *fiber.Ctx) error {
	// get user_id from middleware (stored as string or numeric)
	uidLoc := c.Locals("user_id")
	if uidLoc == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var idUint uint64
	switch v := uidLoc.(type) {
	case string:
		// parse string id
		if parsed, err := strconv.ParseUint(v, 10, 64); err == nil {
			idUint = parsed
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id in token"})
		}
	case uint64:
		idUint = v
	case int:
		idUint = uint64(v)
	case uint:
		idUint = uint64(v)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id type"})
	}

	var user models.User
	if err := database.DB.First(&user, idUint).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	return c.JSON(fiber.Map{
		"id":       user.ID,
		"fullName": user.FullName,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"joined":   user.CreatedAt.Format(time.RFC3339),
	})
}
