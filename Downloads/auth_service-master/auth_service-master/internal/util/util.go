package util

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// -------------------- CONFIG --------------------
var jwtSecret = []byte(getJWTSecret())

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Println("⚠️ JWT_SECRET not set, using default secret (not recommended for production)")
		secret = "super-secret-key-change-this"
	}
	return secret
}

// -------------------- PASSWORD HASH --------------------
func HashPassword(pw string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		log.Panicf("⚠️ failed to hash password: %v", err)
	}
	return string(hash)
}

func CheckPassword(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// -------------------- JWT --------------------
func GenerateJWT(userID, role string, expiryHours int) (string, error) {
	exp := time.Now().Add(time.Duration(expiryHours) * time.Hour)

	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  exp.Unix(),
		"iat":  time.Now().Unix(),
		"iss":  "go-auth-service",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseJWT parses a token string and returns userID and role
func ParseJWT(tokenStr string) (userID, role string, err error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		userID, _ = claims["sub"].(string)
		role, _ = claims["role"].(string)
		return userID, role, nil
	}

	return "", "", fmt.Errorf("invalid token claims")
}

// -------------------- OTP --------------------
func GenerateOTP() string {
	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		log.Panicf("⚠️ failed to generate OTP: %v", err)
	}
	n := int(b[0])<<16 | int(b[1])<<8 | int(b[2])
	return fmt.Sprintf("%06d", n%1000000)
}

// -------------------- EMAIL --------------------
func sendEmail(to, subject, body string) error {
	from := os.Getenv("EMAIL_SENDER")
	pass := os.Getenv("EMAIL_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if from == "" || pass == "" || host == "" || port == "" {
		return fmt.Errorf("email credentials not set in environment variables")
	}

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n\r\n" +
		body + "\r\n")

	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", from, pass, host)

	if err := smtp.SendMail(addr, auth, from, []string{to}, msg); err != nil {
		log.Println("❌ Failed to send email:", err)
		return err
	}

	log.Printf("✅ Email sent to %s (subject: %s)\n", to, subject)
	return nil
}

func SendOTPEmail(toEmail, otp string) error {
	subject := "Your OTP Code"
	body := fmt.Sprintf("Your OTP code is: %s", otp)
	return sendEmail(toEmail, subject, body)
}

func SendResetEmail(toEmail, token string) error {
	subject := "Password Reset Link"
	body := fmt.Sprintf("Reset your password using this token: %s", token)
	return sendEmail(toEmail, subject, body)
}
