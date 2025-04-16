package passwordAuth

import (
	"context"
	"net/http"
	"os"
	"time"

	"go_project/internal/config"
	mongoModel "go_project/internal/models/mongo"
	mysqlModel "go_project/internal/models/mysql"
	"go_project/internal/validators"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordRequest struct {
	OldPassword    string `json:"old_password" binding:"required"`
	NewPassword    string `json:"new_password" binding:"required,min=8,password"`
	RepeatPassword string `json:"repeat_password" binding:"required"`
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", validators.PasswordValidator)
	}
}

func ChangePasswordHandler(ctx *gin.Context) {

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
	var req ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 ตรวจสอบ password และ repeat_password
	if req.NewPassword != req.RepeatPassword {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ New passwords do not match"})
		return
	}

	// 🪸 ตรวจสอบ password เก่า
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Invalid old password"})
		return
	}

	// 🪸 เชื่อมต่อ MongoDB
	db := config.MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := db.Collection("password_history")

	// 🪸 ตรวจสอบ password เก่าที่เคยเปลี่ยน
	var passwordHistoryList []mongoModel.PasswordHistory
	cursor, err := collection.Find(context.Background(), bson.M{"user_id": user.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to fetch password history for checking"})
		return
	}

	// 🪸 ดึงข้อมูลทั้งหมดที่เก็บ
	if err = cursor.All(context.Background(), &passwordHistoryList); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to decode password history for checking"})
		return
	}

	// 🪸 เปรียบเทียบ password เก่าทั้งหมด
	for _, history := range passwordHistoryList {
		err = bcrypt.CompareHashAndPassword([]byte(history.Password), []byte(req.NewPassword))
		if err == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ This password has been used before. Please choose a new one."})
			return
		}
	}

	// 🪸 ปิดข้อมูล
	if err := cursor.Close(context.Background()); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to close cursor"})
		return
	}

	// 🪸 เก็บ password เก่าลง history
	oldHashedPassword := user.Password
	passwordHistory := mongoModel.PasswordHistory{
		UserID:    user.ID,
		Password:  oldHashedPassword,
		ChangedAt: time.Now(),
	}
	_, err = collection.InsertOne(context.Background(), passwordHistory)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to save password history"})
		return
	}

	// 🪸 ลบประวัติ password เก่า ถ้ามีเกิน 5 อัน
	count, err := collection.CountDocuments(context.Background(), bson.M{"user_id": user.ID})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to count password history"})
		return
	}
	if count >= 5 {

		// 🪸 หาและลบ password ที่เก่าที่สุด
		findOptions := options.Find().SetSort(bson.D{{Key: "changed_at", Value: 1}}).SetLimit(1)
		cursorOldest, err := collection.Find(context.Background(), bson.M{"user_id": user.ID}, findOptions)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to find oldest password history"})
			return
		}
		var oldestHistory []mongoModel.PasswordHistory
		if err = cursorOldest.All(context.Background(), &oldestHistory); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to decode oldest password history"})
			return
		}
		if len(oldestHistory) > 0 {
			_, err = collection.DeleteOne(context.Background(), bson.M{"_id": oldestHistory[0].ID})
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to delete oldest password history"})
				return
			}
		}
		if err := cursorOldest.Close(context.Background()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to close cursor"})
			return
		}
	}

	// 🪸 แปลง hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to hash new password"})
		return
	}

	// 🪸 อัพเดท password ใน database
	user.Password = string(hashedPassword)
	if err := config.MySQLClient.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to update password"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 Password changed successfully"})
}
