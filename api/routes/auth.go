package routes

import (
	"go_project/api/handlers/auth"
	adminAuth "go_project/api/handlers/auth/admin"
	emailAuth "go_project/api/handlers/auth/email"
	passwordAuth "go_project/api/handlers/auth/password"
	pictureAuth "go_project/api/handlers/auth/picture"
	tokenAuth "go_project/api/handlers/auth/token"
	"go_project/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(r *gin.Engine) {
	r.POST("/register", auth.RegisterHandler)

	r.POST("/login", middleware.RateLimitMiddleware(middleware.LoginRateLimiter), auth.LoginHandler)

	r.POST("/logout", middleware.Authentication(), auth.LogoutHandler)

	r.POST("/forgot-password", passwordAuth.ForgotPasswordHandler)

	r.POST("/reset-password", middleware.RateLimitMiddleware(middleware.ResetPasswordLimiter), passwordAuth.ResetPasswordHandler)

	r.POST("/change-password", middleware.Authentication(), passwordAuth.ChangePasswordHandler)

	r.POST("/upload-picture", middleware.Authentication(), pictureAuth.UploadPictureHandler)

	adminRoutes := r.Group("/admin")
	{
		adminRoutes.POST("/register", adminAuth.AdminRegisterHandler)
	}

	verifyEmailRoutes := r.Group("/verify-email")
	{
		verifyEmailRoutes.POST("/", middleware.Authentication(), middleware.RateLimitMiddleware(middleware.VerifyEmailLimiter), emailAuth.VerifyEmailHandler)

		verifyEmailRoutes.POST("/confirm", emailAuth.ComfirmVerifyEmailHandler)
	}

	tokenRoutes := r.Group("/token")
	tokenRoutes.Use(middleware.Authentication())
	{
		tokenRoutes.POST("/refresh", middleware.RateLimitMiddleware(middleware.RefreshTokenLimiter), tokenAuth.RefreshHandler)
	}

}
