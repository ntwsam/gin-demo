package backgroundTask

import (
	"log"
	"os"
	"strconv"
	"time"

	"go_project/internal/config"
	mysqlModel "go_project/internal/models/mysql"

	"gorm.io/gorm"
)

func HardDeleteOldUsers(db *gorm.DB, retentionDays int) error {

	// 🪸 ใช้ตรวจหาวันที่ผ่านมาเช่น retentionDays = 30 ก็คือหา 30 วันที่ผ่านมา
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	result := db.Unscoped().
		Where("deleted_at < ?", cutoff).
		Delete(&mysqlModel.Users{})
	return result.Error
}

func HardDeleteOldUsersBackgroundTask(db *gorm.DB) {

	// 🪸 ตั้งให้ทำงานทุก 24 ชม
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	retentionDaysStr := os.Getenv("RETENTION_DAYS")
	if retentionDaysStr == "" {
		retentionDaysStr = "60" // 🐳 ค่า default ถ้าไม่ set
	}

	// 🪸 แปลงให้เป็น int
	retentionDays, err := strconv.Atoi(retentionDaysStr)
	if err != nil {
		log.Printf("⚠️ Invalid RETENTION_DAYS environment variable: %v, using default 60 days", err)
		retentionDays = 60
	}

	for range ticker.C {
		log.Println("⚙️ Running background task: Hard Delete old soft-deleted users")
		err := HardDeleteOldUsers(config.MySQLClient, retentionDays)
		if err != nil {
			log.Printf("⚠️ Error during hard delete: %v", err)
		} else {
			log.Println("✔️ Background task finished: Hard Delete completed")
		}
	}
}
