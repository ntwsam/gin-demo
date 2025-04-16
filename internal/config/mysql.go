package config

import (
	"database/sql"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var MySQLClient *gorm.DB

func ConnectMySQL() {

	// 🪸 ดึงข้อมูลจาก .env
	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbHost := os.Getenv("MYSQL_HOST")
	dbPort := os.Getenv("MYSQL_PORT")
	dbName := os.Getenv("MYSQL_NAME")

	// 🪸 สร้าง data source name - ตัวอย่าง user:password@tcp(host:port)
	dsnWithoutDB := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
	)

	// 🪸 เปิดการเชื่อมต่อ
	db, err := sql.Open("mysql", dsnWithoutDB)
	if err != nil {
		panic(err.Error()) // 🐳 หยุดการทำงานทันทีเมื่อผิดพลาด
	}
	defer db.Close() // 🐳 ปิดการเชื่อมต่อ

	// 🪸 ตรวจเช็คว่ามี db หนือยัง
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)
	_, err = db.Exec(query) // 🐳 สั่งทำงาน query
	if err != nil {
		panic(err.Error())
	}

	// 🪸 สร้าง data source name ใหม่ที่มี database
	dataSourceName := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	// 🪸 เชื่อมต่อ gorm database
	database, err := gorm.Open(mysql.Open(dataSourceName), &gorm.Config{})
	if err != nil {
		panic("⚠️ Failed connecting to MySQL")
	}
	fmt.Println("✔️ MySQL connection successful!")

	MySQLClient = database
}
