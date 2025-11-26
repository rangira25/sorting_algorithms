package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rangira25/auth_service/internal/domain"
	"github.com/rangira25/auth_service/internal/kafka"
	"github.com/rangira25/auth_service/internal/repository"
	"github.com/rangira25/auth_service/internal/util"
	"github.com/redis/go-redis/v9"
)

type AuthService struct {
	rdb      *redis.Client
	repo     *repository.UserRepository
	producer *kafka.KafkaProducer
}

func NewAuthService(rdb *redis.Client, repo *repository.UserRepository, producer *kafka.KafkaProducer) *AuthService {
	return &AuthService{rdb: rdb, repo: repo, producer: producer}
}

// -------------------- USER OPERATIONS --------------------
func (s *AuthService) FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *AuthService) VerifyPassword(input, hash string) bool {
	return util.CheckPassword(input, hash)
}

func (s *AuthService) IssueJWT(user *domain.User) (string, error) {
	return util.GenerateJWT(user.ID, user.Role, 24)
}

func (s *AuthService) HashPassword(password string) string {
	return util.HashPassword(password)
}

func (s *AuthService) CreateUser(ctx context.Context, user *domain.User) error {
	if err := s.repo.Create(ctx, user); err != nil {
		return err
	}

	// Queue welcome email
	s.enqueueNotification(ctx, domain.NotificationPayload{
		Email:   user.Email,
		Subject: "Welcome to Our Platform!",
		Body:    fmt.Sprintf("Hello %s, welcome to the platform!", user.Email),
	})

	return nil
}

// Queue notification for async sending
func (s *AuthService) enqueueNotification(ctx context.Context, payload domain.NotificationPayload) error {
	key := "queue:notifications"
	b, _ := json.Marshal(payload)
	return s.rdb.LPush(ctx, key, b).Err()
}

// -------------------- ADMIN 2FA --------------------
func (s *AuthService) StartAdmin2FA(ctx context.Context, user *domain.User) (string, string, error) {
	otp := util.GenerateOTP()      // generate 6-digit OTP
	tempID := uuid.NewString()     // temporary ID for this OTP session

	if err := s.rdb.Set(ctx, "otp:"+tempID, otp, 5*time.Minute).Err(); err != nil {
		return "", "", fmt.Errorf("failed to store OTP: %w", err)
	}

	if err := s.rdb.Set(ctx, "otp_user:"+tempID, user.Email, 5*time.Minute).Err(); err != nil {
		return "", "", fmt.Errorf("failed to store OTP email mapping: %w", err)
	}

	// Send OTP email
	s.enqueueNotification(ctx, domain.NotificationPayload{
		Email:   user.Email,
		Subject: "Your Admin Login OTP",
		Body:    fmt.Sprintf("Your OTP code is: %s", otp),
	})

	return tempID, otp, nil
}

func (s *AuthService) VerifyAdminOTP(ctx context.Context, tempID, otp string) (string, error) {
	got, err := s.rdb.Get(ctx, "otp:"+tempID).Result()
	if err != nil {
		return "", errors.New("invalid or expired otp")
	}
	if got != otp {
		return "", errors.New("wrong otp")
	}

	email, err := s.rdb.Get(ctx, "otp_user:"+tempID).Result()
	if err != nil {
		return "", errors.New("session expired")
	}

	s.rdb.Del(ctx, "otp:"+tempID, "otp_user:"+tempID)
	return email, nil
}

// -------------------- PASSWORD RESET --------------------
func (s *AuthService) StartPasswordReset(ctx context.Context, user *domain.User) (string, error) {
	code := util.GenerateOTP()

	if err := s.rdb.Set(ctx, "reset:"+code, user.Email, time.Hour).Err(); err != nil {
		return "", fmt.Errorf("failed to store reset code: %w", err)
	}

	// Send reset email
	s.enqueueNotification(ctx, domain.NotificationPayload{
		Email:   user.Email,
		Subject: "Password Reset Code",
		Body:    fmt.Sprintf("Your password reset code is: %s", code),
	})

	return code, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	email, err := s.rdb.Get(ctx, "reset:"+token).Result()
	if err != nil {
		return errors.New("invalid or expired token")
	}
	s.rdb.Del(ctx, "reset:"+token)

	hash := util.HashPassword(newPassword)
	return s.repo.UpdatePassword(ctx, email, hash)
}
