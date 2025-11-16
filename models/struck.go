package models

import "gorm.io/gorm"

type User struct {
    gorm.Model
    FullName string `json:"full_name"`
    Username string `json:"username" gorm:"unique"`
    Email    string `json:"email" gorm:"unique"`
    Password string `json:"-" gorm:"not null"`                         
    Role     string `json:"role" gorm:"type:ENUM('user','instructor','admin');default:'user'"`
}

type Course struct {
    gorm.Model
    Title        string    `json:"title" gorm:"not null"`
    Description  string    `json:"description" gorm:"type:text"` 
    InstructorID uint      `json:"instructor_id"`
    Published    bool      `json:"published" gorm:"default:false"`
    Price        float64   `json:"price" gorm:"default:0"`
    Modules      []Module  `json:"modules" gorm:"constraint:OnDelete:CASCADE"`
    Feedbacks    []Feedback `json:"feedbacks" gorm:"constraint:OnDelete:CASCADE"`
}

type Module struct {
    gorm.Model
    Title    string `json:"title" gorm:"not null"`
    PDFUrl   string `json:"pdf_url"`
    CourseID uint   `json:"course_id"`
    Order    int    `json:"order" gorm:"default:0"`
    Quizzes  []Quiz `json:"quizzes" gorm:"constraint:OnDelete:CASCADE"`
}

type Quiz struct {
    gorm.Model
    ModuleID uint   `json:"module_id"`
    Question string `json:"question" gorm:"not null"`
    Options  string `json:"options" gorm:"type:text"`
    Answer   string `json:"answer" gorm:"not null"`
}

type QuizResult struct {
    gorm.Model
    UserID   uint `json:"user_id"`
    ModuleID uint `json:"module_id"`
    Score    int  `json:"score"`
    Passed   bool `json:"passed"`
}

type Order struct {
    gorm.Model
    UserID   uint    `json:"user_id"`
    CourseID uint    `json:"course_id"`
    Amount   float64 `json:"amount"`
    Status   string  `json:"status" gorm:"default:'pending'"`
    SnapURL  string  `json:"snap_url"`

    User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
    Course Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}

type Enrollment struct {
    gorm.Model
    UserID   uint `json:"user_id"`
    CourseID uint `json:"course_id"`

    User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
    Course Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}

type Feedback struct {
    gorm.Model
    UserID   uint   `json:"user_id"`
    CourseID uint   `json:"course_id"`
    Rating   int    `json:"rating"` // validation in controller
    Comment  string `json:"comment" gorm:"type:text"`

    User   User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
    Course Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}
