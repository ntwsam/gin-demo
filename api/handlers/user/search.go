package user

import (
	"net/http"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"

	"github.com/gin-gonic/gin"
)

func SearchUserHandler(ctx *gin.Context) {

	var users []mysqlModel.Users
	query := config.MySQLClient.Model(&mysqlModel.Users{})

	// 🪸 อ่าน query parameters
	if username := ctx.Query("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%") // แบบมีส่วนใดส่วนนึงตรงกัน
	}
	if email := ctx.Query("email"); email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if status := ctx.Query("status"); status != "" {
		query = query.Where("status = ?", status) // ไม่ใช่ wild card เพราะว่าต้องการเจาะจง
	}
	if role := ctx.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	// 🪸 ทำการ query ข้อมูล
	if err := query.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to search users"})
		return
	}

	// 🪸 ส่งผลลัพท์
	ctx.JSON(http.StatusOK, users)

}
