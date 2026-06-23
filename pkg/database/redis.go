package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Jaruvat303/cashlog/cmd/config"
	"github.com/redis/go-redis/v9"
)

// InitRedisDB ทำหน้าที่เปิดการเชื่อมต่อ Redis และส่งกลับตัวแปร Client กลับไปใช้งาน
func InitRedisDB(cfg *config.Config) *redis.Client {
	// กำหนดออปชั่นในการเชื่อมต่อจากข้อมูลในสัญสักษณ์โครงสร้างคอนฟิก
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost,
		Password: cfg.RedisPassword, // ใส่ค่าว่างหากไม่ได้ตั้งรหัสผ่านไว้
		DB:       cfg.RedisDB,       // โดยปกติเริ่มต้นจะเป็น Database หมายเลข 0

		// ตั้งค่าการจัดการ Connection Pool เบื่องต้น
		PoolSize:        10,
		MinIdleConns:    2,
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ทดสอบการส่งสัญญาน Ping ไปยัง Redis Server
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis server: %v", err)
	}

	fmt.Println("Redis database connection established cleanly via Config Struct")
	return rdb

}
