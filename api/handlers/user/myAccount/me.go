package myAccountUser

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func MyAccountHandler(ctx *gin.Context) {

	// 🪸 ดึง user จาก context
	user, exists := ctx.Get("user")
	if !exists {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ User data not found"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"user": user})
}
