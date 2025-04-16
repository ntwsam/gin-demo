package tokenService

import (
	"context"
	"os"
	"time"

	"go_project/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// 🪸 เจน access token
func GenerateAccessToken(userID string, role string) (string, error) {
	accessTokenClaims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)), // 🐳 หมดอายุใน 15 นาที
			IssuedAt:  jwt.NewNumericDate(time.Now()),                       // 🐳 เวลาที่สร้าง
			NotBefore: jwt.NewNumericDate(time.Now()),                       // 🐳 เวลาที่เริ่มใช้งาน
			Issuer:    "go_project",                                         // 🐳 ผู้สร้าง
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)         // 🐳 สร้างและข้อมูลข้างต้นไว้
	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("JWT_SECRET"))) // 🐳 สร้าง token โดยมี jwt_secret กำกับ
	if err != nil {
		return "", err
	}
	return accessTokenString, nil
}

// 🪸 เจน refresh token
func GenerateRefreshToken(userID string, role string) (string, error) {
	refreshTokenClaims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // 🐳 หมดอายุใน 7 วัน
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "got-rest-api",
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("REFRESH_JWT_SECRET"))) // 🐳 สร้าง token โดยมี refresh_jwt_secret กำกับ
	if err != nil {
		return "", err
	}
	return refreshTokenString, nil
}

// 🪸 เจน access และ refresh token
func GenerateToken(userID string, role string) (string, string, error) {
	accessTokenString, err := GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", err
	}
	refreshTokenString, err := GenerateRefreshToken(userID, role)
	if err != nil {
		return "", "", err
	}
	return accessTokenString, refreshTokenString, nil
}

// 🪸 ตรวจสอบ access token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) { // 🐳 แยก token และ แปลงข้อมูลเป็น claims
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims) // 🐳 ตรวจสอบ claims ที่แยกมา
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// 🪸 ตรวจสอบ refresh token
func ValidateRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("REFRESH_JWT_SECRET")), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// 🪸 เก็บ refresh token ไว้ใน redis
func StoreRefreshToken(userID string, refreshTokenString string) error {
	err := config.RedisClient.Set(context.Background(), userID, refreshTokenString, time.Hour*24*7).Err()
	if err != nil {
		return err
	}
	return nil
}

// 🪸 ดึงข้อมูล refresh token ใน redis
func GetRefreshToken(userID string) (string, error) {
	refreshTokenString, err := config.RedisClient.Get(context.Background(), userID).Result()
	if err != nil {
		return "", err
	}
	return refreshTokenString, nil
}

// 🪸 สร้าง blacklist ไว้เก็บ access token ใน redis
func BlacklistAccessToken(accessTokenString string) error {
	claims := &Claims{}
	_, _, err := new(jwt.Parser).ParseUnverified(accessTokenString, claims) // 🐳 ดึงข้อมูลด้วยไม่ต้องยืนยันเพราะต้องการแค่เวลาหมดอายุ
	if err != nil {
		return err
	}
	expiresAt := claims.ExpiresAt.Unix()
	now := time.Now().Unix()
	expiration := time.Duration(expiresAt-now) * time.Second

	// 🪸 เพิ่ม access token ลงใน blacklist ใน redis
	err = config.RedisClient.Set(context.Background(), accessTokenString, "blacklisted", expiration).Err()
	if err != nil {
		return err
	}
	return nil
}
