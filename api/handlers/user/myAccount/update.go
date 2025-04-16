package myAccountUser

import (
	"net/http"
	"time"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	"go_project/internal/validators"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/ttacon/libphonenumber"
)

type UpdateMyAccountRequest struct {
	Username *string            `json:"username" binding:"omitempty,min=6"`
	Email    *string            `json:"email" binding:"omitempty,email"`
	Phone    *string            `json:"phone" binding:"omitempty,phone"`
	Birthday *string            `json:"birthday" binding:"omitempty"`
	Gender   *mysqlModel.Gender `json:"gender" binding:"omitempty,oneof=female male other prefer_not_say"`
	Role     *mysqlModel.Role   `json:"role" binding:"omitempty,oneof=customer seller"`
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("phone", validators.PhoneValidator)
	}
}

func UpdateMyAccountHandler(ctx *gin.Context) {

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

	// 🪸 แปลง json ให้เป็น struct
	var req UpdateMyAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 สร้าง map เก็บข้อมูลแค่ที่่ต้องการอัพเดท เช่น map["username"]
	updates := make(map[string]interface{})

	// 🪸 ถ้ามี username
	if req.Username != nil && *req.Username != user.Username {

		// 🪸 ตรวจสอบ username และตรวจดูว่าซ้ำมั้ย
		var existingUser mysqlModel.Users
		if err := config.MySQLClient.Where("username = ?", req.Username).First(&existingUser).Error; err == nil && existingUser.ID != user.ID {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ This username is already taken"})
			return
		}
		updates["username"] = req.Username // 🐳 เก็บข้อมูล username
	}
	// 🪸 ถ้ามี email
	if req.Email != nil && *req.Email != user.Email {

		// 🪸 ตรวจสอบ email และตรวจดูว่าซ้ำมั้ย
		var existingUser mysqlModel.Users
		if err := config.MySQLClient.Where("email = ?", req.Email).First(&existingUser).Error; err == nil && existingUser.ID != user.ID {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ This email is already taken"})
			return
		}
		updates["email"] = req.Email // 🐳 เก็บข้อมูล username
	}

	// 🪸 ถ้ามี phone
	if req.Phone != nil {
		num, err := libphonenumber.Parse(*req.Phone, "TH")
		if err != nil {
			formattedNum := libphonenumber.Format(num, libphonenumber.INTERNATIONAL) // 🐳 แปลงไปใช้ country code ในนี้กำหนด th : +66
			updates["phone"] = formattedNum                                          // 🐳 เก็บข้อมูล phone
		}
	}

	// 🪸 ถ้ามี birthday
	if req.Birthday != nil && *req.Birthday != "" {
		parsedBirthday, err := time.Parse("2006-01-02", *req.Birthday) // 🐳 แปลงให้ไปใช้ format YYYY-MM-DD โดย 2006-01-02 เป็น reference time
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Invalid birthday format. Use YYYY-MM-DD"})
			return
		}
		updates["birthday"] = parsedBirthday // 🐳 เก็บข้อมูล birthday
	}

	// 🪸 ถ้ามี gender
	if req.Gender != nil && *req.Gender != "" {
		updates["gender"] = req.Gender
	}

	// 🪸 ถ้ามี role
	if req.Role != nil && *req.Role != "" {
		updates["role"] = mysqlModel.Role(*req.Role)
	}

	// 🪸 ถ้าไม่มีอะไรอัพเดต
	if len(updates) > 0 {
		result := config.MySQLClient.Model(&user).Updates(updates)
		if result.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to update account"})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 Account updated successfully"})
}
