package emailAuth

import (
	"context"
	"net/http"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"

	"github.com/gin-gonic/gin"
)

func ComfirmVerifyEmailHandler(ctx *gin.Context) {

	// 🪸 ดึง token จาก param query
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Invalid verification link"})
		return
	}

	// 🪸 ดึง token verify จาก redis
	userID, err := config.RedisClient.Get(context.Background(), token).Result()
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "⚠️ Invalid or expired token"})
		return
	}

	// 🪸 ดึงข้อมูลจาก db
	var user mysqlModel.Users
	if err := config.MySQLClient.Where("id = ?", userID).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "⚠️ User not found"})
		return
	}

	// 🪸 ตรวจสอบสถานะของ user
	if user.Status == mysqlModel.Active {
		ctx.JSON(http.StatusOK, gin.H{"message": "✔️ Your email has already been verified."})
		return
	} else if user.Status != mysqlModel.Pending {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Invalid user status for verification."})
		return
	}

	// 🪸 เปลี่ยนสถานะให้เป็น active
	user.Status = mysqlModel.Active
	if err := config.MySQLClient.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to update user status"})
		return
	}

	// 🪸 ลบ token ออกจาก redis
	err = config.RedisClient.Del(context.Background(), token).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to delete token from Redis"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 Verify email successfully"})
}
