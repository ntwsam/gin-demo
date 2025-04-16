package user

import (
	"math"
	"net/http"
	"strconv"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"

	"github.com/gin-gonic/gin"
)

func GetAllUsersHanlder(ctx *gin.Context) {

	// 🪸 แปล string เป็น int
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))    // 🐳 ตั้ง default ให้เป็น 1
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10")) // 🐳 ตั้ง default ให้เป็น 10

	var users []mysqlModel.Users
	var total int64

	// 🪸 ตรวจสอบจำนวนแถวทั้งหมดใน database
	config.MySQLClient.Model(&mysqlModel.Users{}).Count(&total)
	offset := (page - 1) * limit // 🐳 ตั้งตำแหน่งเริ่มต้นในข้อมูลที่ต้องการดึงมา

	// 🪸 ดึงข้อมูลทั้งหมดและเก็บไว้ใน users
	if err := config.MySQLClient.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to get users"})
		return
	}

	// 🪸 จำนวนหน้าทั้งหมด
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	ctx.JSON(http.StatusOK, gin.H{
		"users":      users,
		"page":       page,
		"limit":      limit,
		"total":      total,
		"totalPages": totalPages,
	})
}
