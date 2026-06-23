package middleware

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Jaruvat303/cashlog/cmd/config"
	"github.com/Jaruvat303/cashlog/pkg/logger"
	"github.com/Jaruvat303/cashlog/pkg/timeutil"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func NewRequestLogger(cfg *config.Config, baseLogger logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := timeutil.NowInBangkok()

		// ดึง Trace ID จาก Google Cloud Header รูปแบบมักจะเป็น traceid/spanid;o=options)
		gcpTraceHeader := c.Get("X-Cloud-Trace-Context")
		var traceID string

		if gcpTraceHeader != "" {
			parts := strings.Split(gcpTraceHeader, "/")
			if len(parts) > 0 {
				traceID = parts[0]
			}
		}

		if traceID == "" {
			traceID = "local-" + strconv.FormatUint(c.Context().ID(), 10)
		}

		// สร้าง Field สำหรับผูกกับ Logger
		var traceField zap.Field
		if cfg.AppEnv == "production" && !strings.HasPrefix(traceID, "local-") {
			// ฟอร์แมตพิเศษเพื่อให้ Google Cloud Logging เชื่อมโยงเข้า Cloud Trace อัตโนมัติ
			traceField = zap.String("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", cfg.GCProjectID, traceID))
		} else {
			traceField = zap.String("trace_id", traceID)
		}

		// สร้าง Logger ตัวใหม่สำหรับ Request นี้ (สืบทอดมาจากตัวหลัก แต่มี TraceID ติดตัวตลอดไป)
		requestLogger := baseLogger.With(traceField)

		// ใส้ลงไปใน Context และอัปเดต Usecontext ของ Fiber
		ctx := logger.WithContext(c.UserContext(), requestLogger)
		c.SetUserContext(ctx)

		// สั่งให้ Request วิ่งทะลุไปทำงานต่อในเลเยอร์ Handler/Usecase จนเสร็จ
		err := c.Next()

		// หลังจากฝั่งแอปทำงานเสร็จและกำลังจะส่งข้อมูลกลับ -> ทำการบันทึก Log ทันที
		latency := time.Since(start)
		statusCode := c.Response().StatusCode()

		// พ่น Log ทุก Request ด้วย Structured Logging
		requestLogger.Info("HTTP Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
		)

		return err
	}
}
