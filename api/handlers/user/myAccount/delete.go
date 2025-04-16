package myAccountUser

import (
	"context"
	"net/http"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	tokenService "go_project/internal/services/token"

	"github.com/gin-gonic/gin"
)

func DeleteMyAccountHandler(ctx *gin.Context) {

	// 🪸 ดึง user จาก context
	userInterface, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Unauthorized"})
		return
	}

	// 🪸 ดึงข้อมูลในตาราง
	user, ok := userInterface.(mysqlModel.Users)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Internal server error: Invalid user data format"})
		return
	}

	// 🪸 ดึง claims จาก context
	claims, exists := ctx.Get("claims")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to parse claims"})
		return
	}

	// 🪸 แปลงข้อมูล claims
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
	_, err := config.RedisClient.Exists(context.Background(), claimData.UserID).Result()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to check login status"})
		return
	}

	// 🪸 soft delete บัญชี
	result := config.MySQLClient.Delete(&user)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to delete account"})
		return
	}

	// 🪸 ตรวจสอบ ว่าอัพเดตแล้ว
	if result.RowsAffected > 0 {

		// 🪸 ลบ refresh token
		err = config.RedisClient.Del(context.Background(), claimData.UserID).Err()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to delete refresh token"})
			return
		}

		// 🪸 เพิ่ม access token ลงใน black list
		err = tokenService.BlacklistAccessToken(tokenStringData)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to blacklist access token"})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 Account deleted successfully"})

}
