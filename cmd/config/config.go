package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config โครงสร้างข้อมูลหลักที่จะเก็บค่าการตั้งค่าทั้งหมดของ Application
type Config struct {
	AppEnv           string
	Port             string
	APIKey           string
	OwnerNameAliases []string

	// Database Configs
	DBURL             string
	DBMaxIdleConns    int
	DBMaxOpenConns    int
	DBConnMaxLifetime time.Duration

	// Redis Configs
	RedisURL string

	// Google Cloud ProjectID
	GCProjectID string

	// gemini API
	GeminiAPIKey string
	ModelName    string
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
		APIKey:            getEnv("API_KEY", ""), // ปล่อยว่างไว้ หากไม่มีจะไปเช็ต Error ด้านล่าง
		OwnerNameAliases:  getEnvAsStringSlice("OWNER_NAME_ALIASES", []string{"jaruvat", "JARUVAT", "Jaruvat", "จารุุวัฒน์"}),
		DBURL:             getEnv("DB_URL", ""), // ปล่อยว่างไว้ หากไม่มีจะไปเช็ต Error ด้านล่าง
		DBMaxIdleConns:    getEnvAsInt("DB_MAX_IDEL_CONNS", 10),
		DBMaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
		DBConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
		RedisURL:          getEnv("REDIS_URL", ""),
		GCProjectID:       getEnv("GOOGLE_CLOUD_PROJECT", ""),
		GeminiAPIKey:      getEnv("GEMINI_API_KEY", ""),
		ModelName:         getEnv("MODEL_NAME", ""),
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

func getEnvAsStringSlice(key string, defaultVal []string) []string {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}

	var result []string
	if err := json.Unmarshal([]byte(valueStr), &result); err != nil {
		return defaultVal
	}
	return result
}
