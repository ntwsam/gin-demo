package auth

import (
	"net/http"
	"time"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	"go_project/internal/validators"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/ttacon/libphonenumber"
	"golang.org/x/crypto/bcrypt"
)

type RegisgterRequest struct {
	Username       string             `json:"username" binding:"required,min=6"`
	Email          string             `json:"email" binding:"required,email"`
	Password       string             `json:"password" binding:"required,min=8,password"`
	RepeatPassword string             `json:"repeat_password" binding:"required"`
	Phone          *string            `json:"phone" binding:"required,phone"`
	Birthday       string             `json:"birthday" binding:"required"`
	Gender         *mysqlModel.Gender `json:"gender" binding:"required,oneof=female male other prefer_not_say"`
	Role           mysqlModel.Role    `json:"role" binding:"required,oneof=customer merchant"`
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", validators.PasswordValidator)
		v.RegisterValidation("phone", validators.PhoneValidator)
	}
}

func RegisterHandler(ctx *gin.Context) {
	var req RegisgterRequest

	// 🪸 แปลง json ให้เป็น struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 ตรวจ username มีหรือไม่
	var existingUser mysqlModel.Users
	if err := config.MySQLClient.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Username already exists"})
		return
	}

	// 🪸 ตรวจ email มีหรือไม่
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

	// 🪸 แปลง phone ให้เป็น international
	num, err := libphonenumber.Parse(*req.Phone, "TH") // 🐳 แปลงไปใช้ country code ในนี้กำหนด th : +66
	if err != nil {
		formattedNum := libphonenumber.Format(num, libphonenumber.INTERNATIONAL)
		req.Phone = &formattedNum
	}

	// 🪸 แปลง birthday ให้เป็น time.Time
	parseBirthday, err := time.Parse("2006-01-02", req.Birthday) // 🐳 แปลงให้ไปใช้ format YYYY-MM-DD โดย 2006-01-02 เป็น reference time
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Invalid birthday format. Use YYYY-MM-DD"})
		return
	}
	user := mysqlModel.Users{
		ID:       uuid.New().String(),
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
		Phone:    req.Phone,
		Birthday: &parseBirthday,
		Role:     mysqlModel.Role(req.Role),
		Gender:   req.Gender,
	}
	if err := config.MySQLClient.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to create user"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "🎉 User created successfully"})
}
