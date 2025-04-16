package auth

import (
	"net/http"
	"time"

	tokenService "go_project/internal/services/token"

	"github.com/gin-gonic/gin"
)

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RefreshHandler(ctx *gin.Context) {

	// 🪸 แปลง json ให้เป็น struct
	var req RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "⚠️ " + err.Error()})
		println("Error BindJson", err.Error())
		return
	}

	// 🪸 ตรวจสอบ token ใน redis
	claims, err := tokenService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Invalid refresh token"})
		return
	}

	// 🪸 ดึง token ใน redis
	storedRefreshToken, err := tokenService.GetRefreshToken(claims.UserID)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Invalid refresh token"})
		return
	}

	// 🪸 check token
	if storedRefreshToken != req.RefreshToken {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Invalid refresh token"})
		return
	}

	// 🪸 สร้าง token ใหม่
	newAccessTokenString, newRefreshTokenString, err := tokenService.GenerateToken(claims.UserID, claims.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to generate tokens"})
		return
	}

	// 🪸update refresh token ใส่ redis
	err = tokenService.StoreRefreshToken(claims.UserID, newRefreshTokenString)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to update refresh token"})
		return
	}

	// 🪸 อัพเดท cookie ด้วย refresh token ใหม่
	ctx.SetCookie("refresh_token", newRefreshTokenString, int(time.Hour*24*7), "/", "", false, true) // กำหนดอายุ cookie, path, domain, secure, httpOnly

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  newAccessTokenString,
		"refresh_token": newRefreshTokenString,
	})
}
