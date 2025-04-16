package mysqlModel

import (
	"time"

	"gorm.io/gorm"
)

type Gender string

const (
	Female         Gender = "female"
	Male           Gender = "male"
	Other          Gender = "other"
	PreferNotToSay Gender = "prefer not to say"
)

type Role string

const (
	Admin    Role = "admin"
	Customer Role = "customer"
	Merchant Role = "merchant"
)

type Status string

const (
	Active   Status = "active"
	Blocked  Status = "blocked"
	InActive Status = "inactive"
	Pending  Status = "pending"
)

type Users struct {
	ID             string         `gorm:"type:varchar(36);primaryKey;" json:"id"`
	Username       string         `gorm:"unique;not null" json:"username"`
	Email          string         `gorm:"unique;not null" json:"email"`
	Password       string         `gorm:"not null" json:"password"`
	Role           Role           `gorm:"default:'customer'" json:"role"`
	Gender         *Gender        `json:"gender"`
	Phone          *string        `json:"phone"`
	Birthday       *time.Time     `gorm:"type:date" json:"birthday"`
	ProfilePicture *[]byte        `json:"profile_picture"`
	Status         Status         `gorm:"default:'pending'" json:"status"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
