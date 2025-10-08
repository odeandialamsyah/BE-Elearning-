package middlewares

import (
	"backend-elearning/config"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/o1egl/paseto"
)

var v2 = paseto.NewV2()

// AuthMiddleware -> validasi token PASETO
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing token"})
		}

		// accept "Bearer <token>" or raw token
		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		cfg := config.LoadConfig()

		var jsonToken paseto.JSONToken
		var footer string
		if err := v2.Decrypt(token, []byte(cfg.PasetoSecret), &jsonToken, &footer); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token"})
		}

		// Check expiration
		if jsonToken.Expiration.Before(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "token expired"})
		}

		// store common claims in context
		// subject may contain user id
		if jsonToken.Subject != "" {
			c.Locals("user_id", jsonToken.Subject)
		}
		if s := jsonToken.Get("user_id"); s != "" {
			c.Locals("user_id", s)
		}
		if s := jsonToken.Get("role"); s != "" {
			c.Locals("role", s)
		}

		// optional email claim
		if s := jsonToken.Get("email"); s != "" {
			c.Locals("email", s)
		}

		c.Locals("exp", jsonToken.Expiration)

		return c.Next()
	}
}

// RoleMiddleware -> cek role
func RoleMiddleware(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := c.Locals("role")
		if role != requiredRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
		}
		return c.Next()
	}
}
