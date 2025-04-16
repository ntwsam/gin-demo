package user

import (
	"net/http"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	"go_project/internal/validators"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UpdateUserRequest struct {
	Username string          `json:"username" binding:"omitempty,min=6"`
	Email    string          `json:"email" binding:"omitempty,email"`
	Password string          `json:"password" binding:"omitempty,min=8,password"`
	Role     mysqlModel.Role `json:"role" binding:"omitempty,oneof=customer seller"`
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", validators.PasswordValidator)
	}
}

func UpdateUserHandler(ctx *gin.Context) {

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

	// 🪸 แปลง json ให้เป็น struct
	var req UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 อัพเดตเฉพาะที่ต้องการ
	if req.Username != "" {
		user.Username = req.Username
	}

	if req.Email != "" {
		user.Email = req.Email
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to hash password"})
			return
		}
		user.Password = string(hashedPassword)
	}

	if req.Role != "" {
		user.Role = mysqlModel.Role(req.Role)
	}

	if err := config.MySQLClient.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to update user"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 User updated successfully"})
}
