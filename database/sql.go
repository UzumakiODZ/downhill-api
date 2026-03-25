package database

import "time"

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"size:255;unique;not null"` 
	Email    string `gorm:"size:255;unique;not null"`
	RegID    string `gorm:"size:255;unique;not null"`
	Password string `gorm:"size:255;not null"`
}

type Company struct { 
	ID          uint   `gorm:"primaryKey"`
	CompanyName string `gorm:"size:255;unique;not null"`
}

type Role struct {
	ID        uint    `gorm:"primaryKey"`
	RoleName  string  `gorm:"size:255"`
	Year      uint    
	CompanyID uint   
	Company   Company `gorm:"foreignKey:CompanyID;references:ID"`
	OfferType string  `gorm:"size:255"` 
	CGPA      float64 
	Other     string  `gorm:"size:255"`
	CTC       float64
	Base      float64
	Hired     uint
	Converted uint
}

type QuestionBank struct {
	ID        uint    `gorm:"primaryKey"`
	Question  string  `gorm:"type:text"`
	CompanyID uint
	Company   Company `gorm:"foreignKey:CompanyID;references:ID"` 
	Years     uint
}

type Post struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"size:255"`
	Content   string    `gorm:"type:text"` 
	UserID    uint
	User      User      `gorm:"foreignKey:UserID;references:ID"`
	CompanyID  uint
	Company    Company   `gorm:"foreignKey:CompanyID;references:ID"`
	CreatedAt time.Time
}