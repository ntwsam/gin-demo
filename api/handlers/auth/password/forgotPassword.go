package passwordAuth

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

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func ForgotPasswordHandler(ctx *gin.Context) {

	// 🪸 แปลง json ให้เป็น struct
	var req ForgotPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 ตรวจสอบ email
	var user mysqlModel.Users
	if err := config.MySQLClient.Where("email = ?", req.Email).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "⚠️ User not found"})
		return
	}
	// 🪸 สร้าง token สำหรับ reset password
	token := uuid.New().String()

	// 🪸 เก็บ token ใน redis
	err := config.RedisClient.Set(context.Background(), token, user.ID, 15*time.Minute).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to store token in Redis"})
		return
	}

	// 🪸 สร้าง emailData
	baseURL := os.Getenv("BASE_URL")
	resetLink := baseURL + "/reset-password?token=" + token // 🐳 http://localhost:8080/reset-password?token=token
	emailData := email.EmailData{
		To:      []string{user.Email},
		Subject: "Reset Password",
		Body:    "Please click the following to reset your password : " + resetLink,
	}

	// 🪸 ส่ง email
	if err := email.SendEmail(emailData); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to send email"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "📧 Reset password link sent to your email"})
}
