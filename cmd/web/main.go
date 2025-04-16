package main

import (
	"log"
	"os"

	"go_project/api/routes"
	"go_project/internal/config"
	"go_project/internal/middleware"
	"go_project/internal/migrate"
	mysqlModel "go_project/internal/models/mysql"
	"go_project/internal/services"
	"go_project/internal/services/backgroundTask"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {

	// 🪸 โหลด env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("⚠️ Error loading .env file")
	}
}

func main() {

	// 🪸 ดึงไฟล์ใน config มาใช้
	config.ConnectMongoDB()
	config.ConnectMySQL()
	config.ConnectRedis()

	// 🪸 migrate
	migrate.MigrateDatabase(config.MySQLClient, &mysqlModel.Users{})

	// 🪸 router
	r := gin.Default()

	// 🪸 ใช้ cors
	r.Use(middleware.CORS())

	// 🪸 ใช้ routes
	routes.SetupAuthRoutes(r) // 🐳 auth route
	routes.SetupUserRoutes(r) // 🐳 user route

	// 🪸 ใช้ backgroundTask
	go backgroundTask.HardDeleteOldUsersBackgroundTask(config.MySQLClient)

	// 🪸 ตรวจสอบและสร้าง index สำหรับ collection password history
	services.EnsurePasswordHistoryIndex()

	// 🪸 Run Server
	port := os.Getenv("PORT")
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("⚠️ Failed to start server: %v", err)
	}
}
