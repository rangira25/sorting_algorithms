package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rangira25/auth_service/internal/domain"
	"github.com/rangira25/auth_service/internal/service"
	"github.com/rangira25/auth_service/internal/kafka"
)

type AuthHandler struct {
	svc      *service.AuthService
	producer *kafka.KafkaProducer
}

func NewAuthHandler(svc *service.AuthService, producer *kafka.KafkaProducer) *AuthHandler {
	return &AuthHandler{svc: svc, producer: producer}
}

// -------------------- LOGIN --------------------
func (h *AuthHandler) Login(c *gin.Context) {
	var body domain.LoginReq
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.FindUserByEmail(c.Request.Context(), body.Email)
	if err != nil || user == nil || !h.svc.VerifyPassword(body.Password, user.PasswordHash) {
		h.producer.Publish([]byte(fmt.Sprintf(`{"event":"auth.login.failed","email":"%s"}`, body.Email)))
		c.JSON(http.StatusUnauthorized, gin.H{"message": "invalid credentials"})
		return
	}

	if user.Role == "admin" {
		tempID, code, err := h.svc.StartAdmin2FA(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
			return
		}

		h.producer.Publish([]byte(fmt.Sprintf(`{"event":"auth.2fa.required","email":"%s","code":"%s"}`, user.Email, code)))

		c.JSON(http.StatusAccepted, gin.H{"message": "2FA required, OTP sent", "tempId": tempID})
		return
	}

	token, err := h.svc.IssueJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to issue token"})
		return
	}

	h.producer.Publish([]byte(fmt.Sprintf(`{"event":"auth.login.success","email":"%s","user":"%s"}`, user.Email, user.ID)))
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// -------------------- VERIFY 2FA --------------------
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	var req domain.Verify2FAReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email, err := h.svc.VerifyAdminOTP(c.Request.Context(), req.TempID, req.OTP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
		return
	}

	user, err := h.svc.FindUserByEmail(c.Request.Context(), email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "user not found"})
		return
	}

	token, err := h.svc.IssueJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to issue token"})
		return
	}

	h.producer.Publish([]byte(fmt.Sprintf(`{"event":"auth.2fa.success","email":"%s"}`, user.Email)))
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// -------------------- PASSWORD RESET --------------------
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req domain.RequestResetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.svc.FindUserByEmail(c.Request.Context(), req.Email)
	if err == nil && user != nil {
		code, err := h.svc.StartPasswordReset(c.Request.Context(), user)
		if err == nil {
			h.producer.Publish([]byte(fmt.Sprintf(`{"event":"auth.password.reset.requested","email":"%s","code":"%s"}`, user.Email, code)))
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "If the email exists, reset link sent"})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req domain.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	h.producer.Publish([]byte(fmt.Sprintf(`{"event":"auth.password.reset.completed","token":"%s"}`, req.Token)))
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

// -------------------- CREATE USER --------------------
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req domain.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	existing, err := h.svc.FindUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"message": "user already exists"})
		return
	}

	req.PasswordHash = h.svc.HashPassword(req.Password)
	if req.Role == "" {
		req.Role = "user"
	}

	if err := h.svc.CreateUser(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to create user"})
		return
	}

	h.producer.Publish([]byte(fmt.Sprintf(`{"event":"auth.user.created","email":"%s","name":"%s"}`, req.Email, req.Name)))
	c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
}
