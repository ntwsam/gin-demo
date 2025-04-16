package pictureAuth

import (
	"mime/multipart"
	"net/http"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"

	"github.com/gin-gonic/gin"
)

type UploadPictureRequest struct {
	ProfilePicture *multipart.FileHeader `form:"profile_picture" binding:"required"`
}

func UploadPictureHandler(ctx *gin.Context) {

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

	// 🪸 แปลง form-data ให้เป็น struct
	var req UploadPictureRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		return
	}

	// 🪸 เปิดไฟล์
	file, err := req.ProfilePicture.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Unable to open the profile picture. Please check the uploaded data."})
		return
	}
	defer file.Close() // 🐳 ปิดไฟล์

	// 🪸 อ่านไฟล์
	fileBytes := make([]byte, req.ProfilePicture.Size)
	_, err = file.Read(fileBytes)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to read the profile picture. Please try uploading again."})
		return
	}

	// 🪸 ดึงข้อมูลจาก db และ ตรวจสอบ user
	result := config.MySQLClient.Model(&mysqlModel.Users{}).Where("id = ?", user.ID).Update("profile_picture", fileBytes)
	if result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to update profile picture in the database: " + result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "⚠️ User not found or no changes were made"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "🎉 Upload profile picture successfully"})
}
