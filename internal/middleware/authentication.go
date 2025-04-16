package middleware

import (
	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"
	tokenService "go_project/internal/services/token"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 🪸 ดึง token จาก header
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ User not logged in"})
			ctx.Abort()
			return
		}

		// 🪸 ตัด Bearer ออก
		parts := strings.Split(tokenString, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Invalid Authorization header"})
			ctx.Abort()
			return
		}
		tokenString = parts[1]

		// 🪸 ตรวจสอบ access token
		claims, err := tokenService.ValidateToken(tokenString)
		if err != nil {

			// 🪸 ถ้า access token หมดอายุ ตรวจสอบ refresh token ใน cookie
			refreshToken, err := ctx.Cookie("refresh_token")
			if err != nil {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Refresh token not found"})
				ctx.Abort()
				return
			}

			// 🪸 ตรวจสอบ refresh token และสร้าง token ใหม่
			claims, err = tokenService.ValidateRefreshToken(refreshToken)
			if err != nil {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ Invalid refresh token"})
				ctx.Abort()
				return
			}

			newAccessToken, newRefreshToken, err := tokenService.GenerateToken(claims.UserID, claims.Role)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to generate token"})
				return
			}

			// 🪸 เก็บ refresh token ใส่ redis
			err = tokenService.StoreRefreshToken(claims.UserID, newRefreshToken)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "⚠️ Failed to store refresh token"})
				return
			}

			// 🪸 ตั้ง refresh token เป็น cookie
			ctx.SetCookie("refresh_token", newRefreshToken, int(time.Hour*24*7/time.Second), "/", "", false, true)

			// 🪸 ส่ง access token ใหม่ไปยัง header
			ctx.Header("Authorization", "Bearer "+newAccessToken)

			// 🪸 อัปเดตข้อมูล context
			ctx.Set("claims", claims)
			ctx.Set("token", newAccessToken)
		} else {

			// 🪸 ถ้า access token ไม่หมดอายุ ดึง userID จาก claims
			userID := claims.UserID
			var user mysqlModel.Users
			if err := config.MySQLClient.Where("id = ?", userID).First(&user).Error; err != nil {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "⚠️ User not found"})
				ctx.Abort()
				return
			}

			// 🪸 เก็บข้อมูลไว้ใน context
			ctx.Set("claims", claims)
			ctx.Set("token", tokenString)
			ctx.Set("user", user)
		}

		// 🪸 อนุญาตให้ request ไปยัง handler
		ctx.Next()
	}
}
