package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

// POST /checkout
func Checkout(c *fiber.Ctx) error {
	type CheckoutRequest struct {
		CourseID uint `json:"course_id"`
	}

	var req CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	userIDStr := c.Locals("user_id").(string)
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	// Ambil data course
	var course models.Course
	if err := database.DB.First(&course, req.CourseID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "course not found"})
	}

	// Tentukan harga (misal pakai field baru Price)
	price := course.Price
	if price == 0 {
		price = 100000 // default 100 ribu
	}

	// Buat record order di DB (status pending)
	order := models.Order{
		UserID:   uint(userID),
		CourseID: course.ID,
		Amount:   price,
		Status:   "pending",
	}
	if err := database.DB.Create(&order).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Setup Midtrans Snap client
	midclient := snap.Client{}
	midclient.New(os.Getenv("MIDTRANS_SERVER_KEY"), midtrans.Sandbox)

	reqSnap := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  fmt.Sprintf("ORDER-%d", order.ID),
			GrossAmt: int64(price),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			Email: fmt.Sprintf("user%d@example.com", userID),
		},
	}

	// Request Snap Token
	snapResp, err := midclient.CreateTransaction(reqSnap)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Simpan Snap URL ke order
	order.SnapURL = snapResp.RedirectURL
	database.DB.Save(&order)

	return c.JSON(fiber.Map{
		"message":   "checkout created",
		"order_id":  order.ID,
		"snap_url":  snapResp.RedirectURL,
		"amount":    order.Amount,
		"course_id": order.CourseID,
	})
}

// POST /payment/notification
func PaymentNotification(c *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid JSON"})
	}

	orderID := payload["order_id"].(string)
	status := payload["transaction_status"].(string)
	grossAmt := payload["gross_amount"].(string)
	signatureKey := payload["signature_key"].(string)

	// Verify signature
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	hash := sha512.Sum512([]byte(orderID + grossAmt + serverKey))
	expectedSignature := hex.EncodeToString(hash[:])
	if expectedSignature != signatureKey {
		return c.Status(403).JSON(fiber.Map{"error": "invalid signature"})
	}

	// Update status order
	var order models.Order
	if err := database.DB.Where("id = ?", extractOrderID(orderID)).First(&order).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "order not found"})
	}

	order.Status = status
	database.DB.Save(&order)

	// Jika pembayaran sukses, tambahkan ke enrollment
	if status == "settlement" || status == "capture" {
		enrollment := models.Enrollment{
			UserID:   order.UserID,
			CourseID: order.CourseID,
		}
		database.DB.Create(&enrollment)
	}

	return c.JSON(fiber.Map{"message": "payment notification processed"})
}

// GET /me/courses
func GetMyCourses(c *fiber.Ctx) error {
	userIDStr := c.Locals("user_id").(string)
	userID, _ := strconv.ParseUint(userIDStr, 10, 32)

	var enrollments []models.Enrollment
	if err := database.DB.Preload("Course").Where("user_id = ?", userID).Find(&enrollments).Error; err != nil {
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

// Helper untuk ambil ID numerik dari orderID string seperti "ORDER-12"
func extractOrderID(orderID string) uint {
	var id uint
	fmt.Sscanf(orderID, "ORDER-%d", &id)
	return id
}