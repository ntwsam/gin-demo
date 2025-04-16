package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// 🪸 กำหนดอัตราการเรียกของ login
func LoginRateLimiter() *rate.Limiter {
	requestPerSecond := rate.Every(15 * time.Second) // 🐳 กำหนดอนุญาต 1 คำขอทุกๆ 15 วินาที
	burstSize := 3                                   // 🐳 กำหนดร้องขอสูงสุด
	return rate.NewLimiter(requestPerSecond, burstSize)
}

// 🪸 กำหนดอัตราการเรียกของ verify email
func VerifyEmailLimiter() *rate.Limiter {
	requestPerSecond := rate.Every(3 * time.Minute) // 🐳 กำหนดอนุญาต 1 คำขอทุกๆ 5 นาที
	burstSize := 2                                  // 🐳 กำหนดร้องขอสูงสุด
	return rate.NewLimiter(requestPerSecond, burstSize)
}

// 🪸 กำหนดอัตราการเรียกของ verify email
func ResetPasswordLimiter() *rate.Limiter {
	requestPerSecond := rate.Every(3 * time.Minute) // 🐳 กำหนดอนุญาต 1 คำขอทุกๆ 5 นาที
	burstSize := 2                                  // 🐳 กำหนดร้องขอสูงสุด
	return rate.NewLimiter(requestPerSecond, burstSize)
}

// 🪸 กำหนดอัตราการเรียกของ refresh token
func RefreshTokenLimiter() *rate.Limiter {
	requestPerSecond := rate.Every(30 * time.Second) // 🐳 กำหนดอนุญาต 1 คำขอทุกๆ 30 วินาที
	burstSize := 2                                   // 🐳 กำหนดร้องขอสูงสุด
	return rate.NewLimiter(requestPerSecond, burstSize)
}

// 🪸 กำหนดใช้ rate limiter
func RateLimitMiddleware(limiterFunc func() *rate.Limiter) gin.HandlerFunc {
	limiter := limiterFunc()

	return func(ctx *gin.Context) {
		if !limiter.Allow() {
			reservation := limiter.Reserve()
			waitDuration := reservation.Delay()
			reservation.Cancel()
			time := fmt.Sprintf("%d", int(waitDuration.Seconds()))
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "⚠️ Too many requests, Please try again in " + time + " seconds."})
			ctx.Abort() // 🐳 หยุดการประมวลผลต่อไป
			return
		}

		// 🪸 ถ้า request ผ่านทำงานตามปกติ
		ctx.Next()
	}
}
