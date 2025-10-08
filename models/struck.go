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
    Modules     []Module `json:"modules" gorm:"constraint:OnDelete:CASCADE"`
}

type Module struct {
    gorm.Model
    Title    string `json:"title" gorm:"not null"`
    Content  string `json:"content" gorm:"type:text"`
    CourseID uint   `json:"course_id"`
    Order    int    `json:"order" gorm:"default:0"`
}
