package user

import (
	"net/http"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetUserByIDHandler(ctx *gin.Context) {

	// 🪸 รับ user id จาก key parameter url
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Invalid user ID"})
		return
	}

	// 🪸 ตรวจสอบ user
	var user mysqlModel.Users
	if err := config.MySQLClient.Where("id = ?", id).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "⚠️ User not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}
