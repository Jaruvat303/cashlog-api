package database

import (
	"context"
	"time"

	"github.com/Jaruvat303/cashlog/cmd/config"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitRedisDB ทำหน้าที่เปิดการเชื่อมต่อ Redis และส่งกลับตัวแปร Client กลับไปใช้งาน
func InitRedisDB(ctx context.Context, cfg *config.Config, appLogger logger.Logger) *redis.Client {

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		// 🚨 ไม่สั่ง Fatal ให้แอปตาย หากตั้งค่า Redis URL ผิดหรือไม่ได้ตั้งค่าไว้ (ตามหลัก Graceful Degradation)
		appLogger.Warn("⚠️ Failed to parse Redis URL, falling back to default connection options", zap.Error(err))
		opt = &redis.Options{}
	}

	// ตั้งค่า TLS สำหรับการเชื่อมต่อ Redis ผ่าน SSL/TLS เฉพาะกรณีที่ URL ใช้ scheme "rediss://"
	// (ParseURL จะตั้งค่า opt.TLSConfig ให้ไม่เป็น nil เฉพาะตอนที่ URL เป็น rediss:// เท่านั้น)
	if opt.TLSConfig != nil {
		opt.TLSConfig.InsecureSkipVerify = true // ปิดการตรวจสอบใบรับรอง (ไม่แนะนำสำหรับ Production)
	}

	// 🌟 1. เพิ่มการตั้งค่า Connection Pool เพื่อปรับปรุงประสิทธิภาพและความเสถียรของการเชื่อมต่อ

	// 🌟 2. อัดฉีดการตั้งค่า Connection Pool เกรด Senior ของคุณเข้าไปเพิ่ม
	opt.PoolSize = 10
	opt.MinIdleConns = 2
	opt.MaxRetries = 3
	opt.MinRetryBackoff = 8 * time.Millisecond

	// เริ่มต้นเปิดการใช้งาน Client
	rdb := redis.NewClient(opt)

	// สร้าง Context สำหรับดัก Timeout ตอนตรวจสอบสถานะ
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ทดสอบการส่งสัญญาน Ping ไปยัง Redis Server
	_, err = rdb.Ping(pingCtx).Result()
	if err != nil {
		// 🚨 หากเชื่อมต่อไม่ได้บน Production เราจะไม่สั่ง Fatalf จนแอปตาย (ตามหลัก Graceful Degradation)
		// แต่จะบันทึกเป็น WARNING เผื่อให้ระบบถอยไปใช้ฐานข้อมูลตรงๆ แทนได้
		appLogger.Warn("⚠️ Redis connection failed, application will use database fallback", zap.Error(err))
	} else {
		appLogger.Info("⚡ Redis database connection established cleanly via Config Struct!")
	}

	return rdb
}
