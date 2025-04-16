package adminAuth

import (
	"fmt"
	"net/http"
	"os"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	"go_project/internal/validators"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AdminRegisgterRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=8,password"`
	RepeatPassword string `json:"repeat_password" binding:"required"`
	SecretCode     string `json:"secret_code" binding:"required"`
}

func init() {

	// 🪸 ใช้ validator
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", validators.PasswordValidator)
	}
}

func AdminRegisterHandler(ctx *gin.Context) {
	var req AdminRegisgterRequest

	// 🪸 แปลง json ให้เป็น struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 ตรวจสอบ secret code
	secretCode := os.Getenv("ADMIN_SECRET")
	if req.SecretCode != secretCode {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Invalid secret code"})
		return
	}

	// 🪸 ตรวจสอบ email มีหรือไม่
	var existingUser mysqlModel.Users
	if err := config.MySQLClient.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Email already exists"})
		return
	}

	// 🪸 ตรวจสอบ password กับ repeat password ว่าตรงกันหรือไม่
	if req.Password != req.RepeatPassword {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Passwords do not match"})
		return
	}

	// 🪸 เปลี่ยน password เป็น hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost) // 🐳 default : 10
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to hash password"})
		return
	}

	// 🪸 กำหนดชื่อ admin ถ้ามีอยู่แล้วเพิ่มหมายเลข
	adminName := "admin"
	num := 0
	for {
		var adminWithNum string
		if num == 0 {
			adminWithNum = adminName
		} else {
			adminWithNum = fmt.Sprintf("admin%d", num)
		}
		if err := config.MySQLClient.Where("username = ?", adminWithNum).First(&existingUser).Error; err != nil {
			break // 🐳 ถ้าไม่มีข้อผิดพลาด แสดงว่าชื่อไม่ซ้ำ
		}
		num++ // 🐳 ถ้ามีชื่อ admin อยู่แล้ว เพิ่มหมายเลข
	}

	admin := mysqlModel.Users{
		ID:       uuid.New().String(),
		Username: adminName,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     mysqlModel.Admin,
		Status:   mysqlModel.Active,
	}
	if err := config.MySQLClient.Create(&admin).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to create admin"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "🎉 Admin created successfully"})
}
