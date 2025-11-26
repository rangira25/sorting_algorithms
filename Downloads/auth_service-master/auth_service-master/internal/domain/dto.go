package domain

import (
	"time"

	"gorm.io/gorm"
)

// User 
type User struct {
	ID           string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email        string         `gorm:"type:citext;uniqueIndex;not null" json:"email"`
	Password     string         `gorm:"-" json:"password,omitempty"`      
	PasswordHash string         `gorm:"type:text;not null" json:"-"`
	Role         string         `gorm:"type:varchar(32);default:'user'" json:"role"` // admin or user
	Status       string         `gorm:"type:varchar(16);default:'active'" json:"status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
		Name         string 
}

// -------------------- DTOs --------------------

// LoginReq is the payload for login
type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Verify2FAReq for admin 2FA
type Verify2FAReq struct {
	TempID string `json:"tempId" binding:"required"`
	OTP    string `json:"otp" binding:"required"`
}

// RequestResetReq to request password reset
type RequestResetReq struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordReq for resetting password
type ResetPasswordReq struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}
