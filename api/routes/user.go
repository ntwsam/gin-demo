package routes

import (
	"go_project/api/handlers/user"
	myAccountUser "go_project/api/handlers/user/myAccount"
	"go_project/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(r *gin.Engine) {

	userRoutes := r.Group("/users")
	userRoutes.Use(middleware.Authentication(), middleware.Authorization([]string{"admin"}))
	{
		userRoutes.GET("/", user.GetAllUsersHanlder)               // 🐳 มีการใช้ Query Param : page กับ limit
		userRoutes.GET("/:id", user.GetUserByIDHandler)            // 🐳 ใช้ Path Param
		userRoutes.PUT("/:id", user.UpdateUserHandler)             // 🐳 ใช้ Path Param
		userRoutes.DELETE("/:id", user.DeleteUserHandler)          // 🐳 ใช้ Path Param
		userRoutes.GET("/get-picture/:id", user.GetProfilePicture) // 🐳 ใช้ Path Param
		userRoutes.PUT("/status/:id", user.ManageBlockUserHandler) // 🐳 ใช้ Path Param
		userRoutes.GET("/search", user.SearchUserHandler)          // 🐳 มีการใช้ Query Param : username, email, status และ role
	}

	meRoute := r.Group("/myProfile")
	meRoute.Use(middleware.Authentication())
	{
		meRoute.GET("/", myAccountUser.MyAccountHandler)
		meRoute.PUT("/", myAccountUser.UpdateMyAccountHandler)
		meRoute.DELETE("/", myAccountUser.DeleteMyAccountHandler)
	}
}
