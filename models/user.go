package models

import (
	"github.com/jinzhu/gorm"
)

// User model contains all relevant details of a particular user
type User struct {
	gorm.Model
	Token string
	Name  string
	Email string
}