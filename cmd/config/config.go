package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config โครงสร้างข้อมูลหลักที่จะเก็บค่าการตั้งค่าทั้งหมดของ Application
type Config struct {
	AppEnv string
	Port   string

	// Database Configs
	DBURL             string
	DBMaxIdleConns    int
	DBMaxOpenConns    int
	DBConnMaxLifetime time.Duration

	// Redis Configs
	RedisHost     string
	RedisPassword string
	RedisDB       int

	// Google Cloud ProjectID
	GCProjectID string
}

// LoadConfig ทำหน้าที่โหลดไฟล์ .env และแปลงค่ามาใส่ใน struct Config
func LoadConfig() *Config {
	// โหลดไฟล์ .env สำหรับการทำงาน Local (หากไม่เจอก็จะไปดึงจาก Environment ของระบบ/Docker แทน)
	if err := godotenv.Load(); err != nil {
		log.Println("Note: .env file not found, loading options from OS environment")
	}

	return &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		Port:              getEnv("PORT", "8080"),
		DBURL:             getEnv("DB_URL", ""), // ปล่อยว่างไว้ หากไม่มีจะไปเช็ต Error ด้านล่าง
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDEL_CONNS", 10),
		DBMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
		DBConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		RedisHost:         getEnv("REDIS_HOST", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           getEnvAsInt("REDIS_DB", 0),
		GCProjectID:       getEnv("GOOGLE_CLOUD_PROJECT", ""),
	}
}

// function Utility สำหรับดึงค่าและกำหนดค่าเริ่มต้น
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}

	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if d, err := time.ParseDuration(valueStr); err == nil {
		return d
	}

	return defaultVal
}
