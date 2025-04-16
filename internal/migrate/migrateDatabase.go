package migrate

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

func MigrateDatabase(db *gorm.DB, model ...interface{}) {
	if err := db.AutoMigrate(model...); err != nil {
		log.Fatalf("⚠️ Failed to migrate database: %v", err)
	}
	fmt.Println("✔️ Database migration successful!")
}
