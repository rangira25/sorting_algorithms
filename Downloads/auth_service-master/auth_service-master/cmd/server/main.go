package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rangira25/auth_service/internal/config"
	"github.com/rangira25/auth_service/internal/handlers"
	"github.com/rangira25/auth_service/internal/kafka"
	"github.com/rangira25/auth_service/internal/repository"
	"github.com/rangira25/auth_service/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// -------------------- LOAD .ENV --------------------
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// -------------------- REDIS --------------------
	rdb := config.NewRedisClient() // reads REDIS_ADDR, REDIS_PASSWORD, REDIS_DB

	// -------------------- DATABASE --------------------
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSL := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		dbHost, dbUser, dbPassword, dbName, dbPort, dbSSL,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database:", err)
	}

	// -------------------- REPOSITORY --------------------
	userRepo := repository.NewUserRepository(db)

	// -------------------- KAFKA PRODUCER --------------------
	kafkaBrokers := []string{os.Getenv("KAFKA_BROKER")} // e.g., "localhost:9092"
	kafkaTopic := "auth-events"
	producer, err := kafka.NewKafkaProducer(kafkaBrokers, kafkaTopic)
	if err != nil {
		log.Fatal("failed to create kafka producer:", err)
	}
	defer producer.Close()

	// -------------------- SERVICE --------------------
	authSvc := service.NewAuthService(rdb, userRepo, producer)

	// -------------------- HANDLER --------------------
	handler := handlers.NewAuthHandler(authSvc, producer)

	// -------------------- ROUTER --------------------
	r := gin.Default()
	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", handler.Login)
		v1.POST("/auth/verify-2fa", handler.Verify2FA)
		v1.POST("/auth/request-reset", handler.RequestPasswordReset)
		v1.POST("/auth/reset", handler.ResetPassword)
		v1.POST("/auth/register", handler.CreateUser)
	}

	// -------------------- SERVER --------------------
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Println("Server running on http://localhost:8080")
	log.Fatal(srv.ListenAndServe())
}
