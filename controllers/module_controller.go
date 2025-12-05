package controllers

import (
	"backend-elearning/database"
	"backend-elearning/models"
	"fmt"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// AddModuleToCourse -> POST /instructor/courses/:id/modules (instructor only)
func AddModuleToCourse(c *fiber.Ctx) error {
    courseID := c.Params("course_id")

    // Ambil form-data
    title := c.FormValue("title")
    orderStr := c.FormValue("order")

    if title == "" {
        return c.Status(400).JSON(fiber.Map{"error": "title is required"})
    }

    // Konversi order ke int
    order, _ := strconv.Atoi(orderStr)

    // Cek course exist
    var course models.Course
    if err := database.DB.First(&course, courseID).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "course not found"})
    }

    // Ambil file PDF (opsional)
    file, _ := c.FormFile("pdf")

    // Buat module dulu
    module := models.Module{
        Title:    title,
        Order:    order,
        CourseID: course.ID,
    }

    if err := database.DB.Create(&module).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // Jika tidak ada file PDF → return module tanpa PDF
    if file == nil {
        return c.Status(201).JSON(fiber.Map{
            "message": "module created (without PDF)",
            "module":  module,
        })
    }

    // Validasi PDF
    if file.Header.Get("Content-Type") != "application/pdf" {
        return c.Status(400).JSON(fiber.Map{"error": "file must be PDF"})
    }

    // Tentukan lokasi penyimpanan
    filename := fmt.Sprintf("uploads/modules/%d-%s", module.ID, file.Filename)

    if err := c.SaveFile(file, filename); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // Update module dengan PDFUrl
    module.PDFUrl = "/" + filename
    database.DB.Save(&module)

    return c.Status(201).JSON(fiber.Map{
        "message": "module created successfully",
        "module":  module,
    })
}

// EditModule -> PUT /instructor/courses/:course_id/modules/:id (requires instructor)
func EditModule(c *fiber.Ctx) error {
    courseID := c.Params("course_id")
    moduleID := c.Params("module_id")

    title := c.FormValue("title")
    orderStr := c.FormValue("order")
    order, _ := strconv.Atoi(orderStr)

    var module models.Module
    // Ensure module exists and belongs to the given course
    if err := database.DB.Where("id = ? AND course_id = ?", moduleID, courseID).First(&module).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "module not found for this course"})
    }

    // Update Title & Order
    if title != "" {
        module.Title = title
    }
    module.Order = order

    // Cek PDF baru (opsional)
    pdfFile, _ := c.FormFile("pdf")
    if pdfFile != nil {
        if pdfFile.Header.Get("Content-Type") != "application/pdf" {
            return c.Status(400).JSON(fiber.Map{"error": "file must be a PDF"})
        }

        // Hapus PDF lama jika ada
        if module.PDFUrl != "" {
            _ = os.Remove("." + module.PDFUrl)
        }

        // Simpan PDF baru
        filename := fmt.Sprintf("uploads/modules/%d-%s", module.ID, pdfFile.Filename)
        if err := c.SaveFile(pdfFile, filename); err != nil {
            return c.Status(500).JSON(fiber.Map{"error": err.Error()})
        }

        module.PDFUrl = "/" + filename
    }

    if err := database.DB.Save(&module).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "message": "module updated successfully",
        "module":  module,
    })
}

// DeleteModule -> DELETE /instructor/courses/:course_id/modules/:id (requires instructor)
func DeleteModule(c *fiber.Ctx) error {
    courseID := c.Params("course_id")
    id := c.Params("module_id")

    var module models.Module
    if err := database.DB.Where("id = ? AND course_id = ?", id, courseID).First(&module).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "module not found for this course"})
    }

	if err := database.DB.Delete(&module).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
        "message": "module deleted successfully",
    })

}