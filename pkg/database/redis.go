package database

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/Jaruvat303/cashlog/cmd/config"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// InitRedisDB ทำหน้าที่เปิดการเชื่อมต่อ Redis และส่งกลับตัวแปร Client กลับไปใช้งาน
func InitRedisDB(ctx context.Context, cfg *config.Config) *redis.Client {
	var opt *redis.Options
	var err error

	opt = &redis.Options{
		Addr:     cfg.RedisHost,
		Username: cfg.RedisUsename,
		Password: cfg.RedisPassword,

		// สำคัญมากสำหรับ Upstash Cloud: ต้องเปิด TLS Configuration ด้วยถ้าใช้คลาวด์
		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

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
		logger.Ctx(ctx).Warn("⚠️ Redis connection failed, application will use database fallback", zap.Error(err))
	} else {
		logger.Ctx(ctx).Info("⚡ Redis database connection established cleanly via Config Struct!")
	}

	return rdb
}
