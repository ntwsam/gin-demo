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

type ResetPasswordRequest struct {
	Password       string `json:"password" binding:"required,min=8,password"`
	RepeatPassword string `json:"repeat_password" binding:"required"`
}

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("password", validators.PasswordValidator)
	}
}

func ResetPasswordHandler(ctx *gin.Context) {

	// 🪸 ดึง token จาก query parameter
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Token is required"})
		return
	}

	// 🪸 แปลง json ให้เป็น struct
	var req ResetPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 ดึง userID จาก redis ด้วย token
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

	// 🪸 ตรวจสอบ password และ repeat_password
	if req.Password != req.RepeatPassword {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ New passwords do not match"})
		return
	}

	// 🪸 เชื่อมต่อ MongoDB
	db := config.MongoClient.Database(os.Getenv("MONGO_DATABASE"))
	collection := db.Collection("password_history")

	// 🪸 ตรวจสอบ password เก่า
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
		err = bcrypt.CompareHashAndPassword([]byte(history.Password), []byte(req.Password))
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
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

	// 🪸 ลบ token ใน redis
	err = config.RedisClient.Del(context.Background(), token).Err()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to delete token from Redis"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 Password reset successfully"})
}
