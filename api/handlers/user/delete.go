package user

import (
	"context"
	"net/http"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	tokenService "go_project/internal/services/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func DeleteUserHandler(ctx *gin.Context) {

	// 🪸 รับ user id จาก key parameter url
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Invalid user ID"})
		return
	}

	// 🪸 ตรวจสอบ user
	var user mysqlModel.Users
	if err := config.MySQLClient.First(&user, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "⚠️ User not found"})
		return
	}

	// 🪸 กรณีผู้ใช้ยัง login ดึง access token จาก context
	tokenString, exists := ctx.Get("token")
	if exists {
		tokenStringData, ok := tokenString.(string)
		if ok {

			// 🪸 เพิ่ม token ลง blacklist
			err = tokenService.BlacklistAccessToken(tokenStringData)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to blacklist access token"})
				return
			}
		}
	}

	// 🪸 ลบ refresh token
	err = config.RedisClient.Del(context.Background(), user.ID).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to delete refresh token"})
		return
	}

	// 🪸 ลบ user จาก database
	if err := config.MySQLClient.Delete(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to delete user"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 User deleted successfully"})

}
