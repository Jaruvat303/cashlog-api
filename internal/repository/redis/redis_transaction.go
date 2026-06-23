package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisDashboardRepository struct {
	rdb *redis.Client
}

const FileCashePrefix = "processed_file: %s"

func NewRedisDashboardRepository(rdb *redis.Client) *redisDashboardRepository {
	return &redisDashboardRepository{rdb: rdb}
}

// GetCache implements [domain.DashboardRepository].
func (r *redisDashboardRepository) GetCache(ctx context.Context, periodKey string) (string, error) {
	return r.rdb.Get(ctx, "dashboard:"+periodKey).Result()
}

// SetCache implements [domain.DashboardRepository].
func (r *redisDashboardRepository) SetCache(ctx context.Context, periodKey string, jsonData string) error {
	return r.rdb.Set(ctx, "dashboard:"+periodKey, jsonData, 1*time.Hour).Err()
}

func (r *redisDashboardRepository) InvalidateCache(ctx context.Context, periodKey string) error {
	return r.rdb.Del(ctx, "dashboard:"+periodKey).Err()
}

// ChackFlieExist ตรวจสอบชื่อไฟล์่ภาพเคยประมาลผลไปแล้วรึยัง
func (r *redisDashboardRepository) CheckFileExists(ctx context.Context, localImageName string) (bool, error) {
	key := fmt.Sprintf(FileCashePrefix, localImageName)

	val, err := r.rdb.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	// หาก key มีค่า val จะเท่ากับ 1
	return val > 0, nil
}

// SetFlieCache สั่งจำชื่อไฟล์สลืปใบลงใน Redis และกำหนดเวลาหมดอายุ TTL ไว้ที่ 30 วัน
func (r *redisDashboardRepository) SetFileCache(ctx context.Context, localImageName string) error {
	key := fmt.Sprintf(FileCashePrefix, localImageName)

	// บันทึกไว้ 30 วัน (เท่ากับกรอบเวลาที่ flutter เช็ตภาพย้อนหลัง)
	ttl := 30 * 24 * time.Hour

	return r.rdb.Set(ctx, key, "true", ttl).Err()
}
