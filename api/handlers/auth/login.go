package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	tokenService "go_project/internal/services/token"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"password"`
}

type LoginResponse struct {
	Message      string `json:"message"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func LoginHandler(ctx *gin.Context) {

	// 🪸 แปลง json ให้เป็น struct
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 ตรวจสอบ username หรือ email
	var user mysqlModel.Users
	if err := config.MySQLClient.Where("username = ? OR email = ?", req.UsernameOrEmail, req.UsernameOrEmail).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Invalid username or email"})
		return
	}

	// 🪸 ตรวจสอบ password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Invalid password"})
		return
	}

	// 🪸 ตรวจสอบผู้ใช้ ใช้งานอยู่รึมั้ย
	exists, err := config.RedisClient.Exists(context.Background(), user.ID).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to check login status"})
		return
	}
	if exists == 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ User already logged in"})
		return
	}

	// 🪸 สร้าง JWT token
	accessTokenString, refreshTokenString, err := tokenService.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to generate token"})
		return
	}

	// 🪸 เก็บ refresh token ใส่ redis
	err = tokenService.StoreRefreshToken(user.ID, refreshTokenString)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to store refresh token"})
		return
	}

	// 🪸 ตั้ง refresh token เป็น cookie
	ctx.SetCookie("refresh_token", refreshTokenString, int(time.Hour*24*7/time.Second), "/", "", false, true)

	// 🪸 ตั้ง access token ใน authorization header
	ctx.Header("Authorization", "Bearer "+accessTokenString)

	ctx.JSON(http.StatusOK, LoginResponse{
		Message:      fmt.Sprintf("🎉 Welcome back %s", user.Username),
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	})

}
