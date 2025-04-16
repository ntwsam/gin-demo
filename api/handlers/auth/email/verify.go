package emailAuth

import (
	"context"
	"net/http"
	"os"
	"time"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	"go_project/internal/services/email"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func VerifyEmailHandler(ctx *gin.Context) {

	// 🪸 ดึง user จาก context
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Unauthorized"})
		return
	}
	user, ok := userInterface.(mysqlModel.Users)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Internal server error: Invalid user data format"})
		return
	}

	// 🪸 สร้าง token ใหม่
	token := uuid.New().String()

	// 🪸 สร้าง value แล้วเก็บ token ไว้ใน redis
	err := config.RedisClient.Set(context.Background(), token, user.ID, 15*time.Minute).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to store token in Redis"})
		return
	}

	// 🪸 สร้าง emailData
	baseURL := os.Getenv("BASE_URL")
	resetLink := baseURL + "/verify-email/confirm?token=" + token // 🐳 http://localhost:8080/verify-email/confirm?token=token
	emailData := email.EmailData{
		To:      []string{user.Email},
		Subject: "Verify Email",
		Body:    "Please click the following to verify your email : " + resetLink,
	}

	// 🪸 ส่ง email
	if err := email.SendEmail(emailData); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to send email"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "📧 Verify Email link sent to your email"})
}
