package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FullName string `json:"full_name"`
	Username string `json:"username" gorm:"unique"`
	Email    string `json:"email" gorm:"unique"`
	Password 	string `json:"password" gorm:"not null"`
	Role     string `json:"role"` // user, instructor, admin
}

type Course struct {
    gorm.Model
    Title       string   `json:"title" gorm:"not null"`
    Description string   `json:"description" gorm:"type:text"`
    InstructorID uint    `json:"instructor_id"`
    Published   bool     `json:"published" gorm:"default:false"`
	Price        float64  `json:"price" gorm:"default:0"`
    Modules     []Module `json:"modules" gorm:"constraint:OnDelete:CASCADE"`
}

type Module struct {
    gorm.Model
    Title    string `json:"title" gorm:"not null"`
    Content  string `json:"content" gorm:"type:text"`
    CourseID uint   `json:"course_id"`
    Order    int    `json:"order" gorm:"default:0"`
}

type Quiz struct {
	gorm.Model
	ModuleID uint   `json:"module_id"`
	Question string `json:"question" gorm:"not null"`
	Options  string `json:"options" gorm:"type:text"` // JSON array of options
	Answer   string `json:"answer" gorm:"not null"`   // Correct answer
}

type QuizResult struct {
	gorm.Model
	UserID   uint  `json:"user_id"`
	ModuleID uint  `json:"module_id"`
	Score    int   `json:"score"`
	Passed   bool  `json:"passed"`
}

type Order struct {
	gorm.Model
	UserID   uint    `json:"user_id"`
	CourseID uint    `json:"course_id"`
	Amount   float64 `json:"amount"`
	Status   string  `json:"status" gorm:"default:'pending'"` // pending, paid, failed, cancelled
	SnapURL  string  `json:"snap_url"`
	User     User    `gorm:"foreignKey:UserID"`
	Course   Course  `gorm:"foreignKey:CourseID"`
}

// Enrollment merepresentasikan user yang sudah terdaftar di course
type Enrollment struct {
	gorm.Model
	UserID   uint   `json:"user_id"`
	CourseID uint   `json:"course_id"`
	User     User   `gorm:"foreignKey:UserID"`
	Course   Course `gorm:"foreignKey:CourseID"`
}