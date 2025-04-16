package user

import (
	"fmt"
	"go_project/internal/config"
	mongoModel "go_project/internal/models/mongo"
	mysqlModel "go_project/internal/models/mysql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func ManageBlockUserHandler(ctx *gin.Context) {

	// 🪸 รับ user id จาก key parameter url
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ Invalid user ID"})
		return
	}

	// 🪸 ตรวจสอบ user
	var user mysqlModel.Users
	if err := config.MySQLClient.Model(user).Where("id = ?", id).First(&user).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "⚠️ User not found"})
		return
	}

	isBlocked := user.Status == mysqlModel.Blocked
	var newStatus mysqlModel.Status

	mongoDatabase := os.Getenv("MONGO_DATABASE")

	// 🪸 status เป้น blocked
	if isBlocked {

		// 🪸 ดึงสถานะก่อนหน้านี้
		collection := config.MongoClient.Database(mongoDatabase).Collection("status_history")
		filter := bson.M{"user_id": user.ID}
		var history mongoModel.StatusHistory
		err := collection.FindOne(ctx, filter).Decode(&history)
		if err == nil {
			newStatus = mysqlModel.Status(history.Status)

			// 🪸 ลบประวัติ
			_, err := collection.DeleteOne(ctx, bson.M{"_id": history.ID})
			if err != nil {
				log.Printf("⚠️ Error deleting status history from MongoDB: %v", err)
			}
		} else {
			newStatus = mysqlModel.Pending
		}
	} else {

		collection := config.MongoClient.Database(mongoDatabase).Collection("status_history")
		_, err := collection.InsertOne(ctx, mongoModel.StatusHistory{
			UserID: user.ID,
			Status: string(user.Status),
		})
		if err != nil {
			log.Printf("⚠️ Error saving status history to MongoDB: %v", err)
		}
		newStatus = mysqlModel.Blocked
	}

	// 🪸 อัพเดต status
	user.Status = newStatus
	if err := config.MySQLClient.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to update user status"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("🎉 User status updated to: %s", user.Status)})
}
