package middleware

import (
	"net/http"
	"strings"

	tokenService "go_project/internal/services/token"

	"github.com/gin-gonic/gin"
)

func Authorization(roles []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 🪸 ดึงข้อมูล user จาก context
		user, exists := ctx.Get("claims")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "⚠️ Unauthorized: User not authenticated"})
			ctx.Abort()
			return
		}

		// 🪸 แปลงข้อมูล user
		userData, ok := user.(*tokenService.Claims)
		if !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "⚠️ Unauthorized: User not authenticated"})
			ctx.Abort()
			return
		}

		// 🪸 ตรวจสอบ role
		if userData.Role == "" {
			ctx.JSON(http.StatusForbidden, gin.H{"message": "⚠️ Unauthorized: User role not defined"})
			ctx.Abort()
			return
		}

		// 🪸 ตรวจสอบว่า role ของ user อยู่ใน slice ของ roles ที่อนุญาตหรือไม่
		allowed := false
		for _, role := range roles {
			if strings.EqualFold(userData.Role, role) { // 🐳 เปรียบเทียบกันโดยไม่สนว่าจะพิมพ์เล็กพิมพ์ใหญ่
				allowed = true
				break
			}
		}

		if !allowed {
			ctx.JSON(http.StatusForbidden, gin.H{"message": "⚠️ Unauthorized: Insufficient privileges"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
