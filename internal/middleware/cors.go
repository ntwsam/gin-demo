package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {

	// 🪸 ดึงข้อมูลจาก .env
	allowedOrigin := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigin == "" {
		allowedOrigin = "*" // 🐳 default
	}

	// 🪸 แยกค่า
	originLists := strings.Split(allowedOrigin, ",")

	return func(ctx *gin.Context) {

		// 🪸 ดึง origin จาก request header
		origin := ctx.Request.Header.Get("Origin")
		if origin != "" {
			allowed := false
			for _, allallowedOrigin := range originLists {

				// 🪸 ลบช่องว่างหน้าหลังของ string
				trimmedOrigin := strings.TrimSpace(allallowedOrigin)
				if trimmedOrigin == "*" || origin == trimmedOrigin {
					allowed = true
					break
				}
			}
			if allowed {
				ctx.Header("Access-Control-Allow-Origin", origin)                             // 🐳 กำหนดให้โดเมนใดที่เข้าถึง resources ได้บ้าง)
				ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // 🐳 กำหนดให้ http อนุญาต GET POST PUT DELETE OPTIONS
				ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")     // 🐳 กำหนดให้ header อนุญาต content-type และ Authorization
				ctx.Header("Access-Control-Allow-Credentials", "true")                        // 🐳 กำหนดให้ส่ง cookie ข้ามโดเมนได้

				// 🪸 หากเป็น options ยุติประมวลผล
				if ctx.Request.Method == "OPTIONS" {
					ctx.AbortWithStatus(http.StatusNoContent)
					return
				}
			}
		}
		// 🪸 หากไม่เป็น options ก็ไปขั้นต่อไป
		ctx.Next()
	}
}
