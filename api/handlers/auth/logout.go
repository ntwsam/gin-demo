package auth

import (
	"context"
	"net/http"

	"go_project/internal/config"
	tokenService "go_project/internal/services/token"

	"github.com/gin-gonic/gin"
)

func LogoutHandler(ctx *gin.Context) {

	// 🪸 ดึง claims จาก context
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to get claims"})
		return
	}

	// 🪸 แปลงข้อมูล
	claimData, ok := claims.(*tokenService.Claims)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to parse claims"})
		return
	}

	// 🪸 ดึง access token จาก context
	tokenString, exists := ctx.Get("token")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to get token"})
		return
	}

	// 🪸 แปลงข้อมูล token
	tokenStringData, ok := tokenString.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to parse token"})
		return
	}

	// 🪸 ตรวจสอบว่ามี refresh token อยู่ใน redis หรือไม่
	count, err := config.RedisClient.Exists(context.Background(), claimData.UserID).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to check login status"})
		return
	}
	if count == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ User not logged in"})
		return
	}

	// 🪸 ลบ refresh token ออกจาก Redis
	err = config.RedisClient.Del(context.Background(), claimData.UserID).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to delete refresh token"})
		return
	}

	// 🪸 เพิ่ม access token ลงใน blacklist
	err = tokenService.BlacklistAccessToken(tokenStringData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to blacklist access token"})
		return
	}
	// 🪸 ลบ cookie
	ctx.SetCookie("refresh_token", "", 0, "/", "", false, true)

	// 🪸 ลบ authorization header
	ctx.Header("Authorization", "")

	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 Logout successful"})
}
