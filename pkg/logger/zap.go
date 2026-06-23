package logger

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey struct{}

// Logger คือ Interface ที่เราจะนำไปใช้ประกาศใน Layer อื่น ๆ (เช่น Usecase)
// เพื่อลดการผูกมัดกับโครงสร้างของ zap โดยตรง (Loose Coupling)
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
}

// zapLogger คือ struct ที่ implement Logger interface ข้างบน
type zapLogger struct {
	util *zap.Logger
}

// With ช่วยให้เราสามารถเพิ่มฟิลด์ถาวรเข้าไปใน Logger ตัวนั้นได้ (เช่น Trace ID)
func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{util: l.util.With(fields...)}
}

func (l *zapLogger) Debug(msg string, fields ...zap.Field) { l.util.Debug(msg, fields...) }
func (l *zapLogger) Info(msg string, fields ...zap.Field)  { l.util.Info(msg, fields...) }
func (l *zapLogger) Warn(msg string, fields ...zap.Field)  { l.util.Warn(msg, fields...) }
func (l *zapLogger) Error(msg string, fields ...zap.Field) { l.util.Error(msg, fields...) }
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) { l.util.Fatal(msg, fields...) }

func InitLogger(env string) Logger {
	var config zap.Config

	if env == "production" {
		// ฟอร์แมต JSON สำหรับ Google Cloud Logging
		config = zap.NewProductionConfig()
		config.EncoderConfig.LevelKey = "severity"
		config.EncoderConfig.EncodeLevel = gcpLevelEncoder
	} else {
		// ฟอร์แมตอ่านง่ายสำหรับเครื่อง Local
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// สั่งข้าม Caller 1 ระดับเพื่อให้ Log แสดงบรรทัดที่เรียกใช้จริง ๆ ไม่ใช่บรรทัดใน wrapper นี้
	zLog, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}

	return &zapLogger{util: zLog}
}

// NewNopLogger สำหรับใช้ Unit test
func NewNopLogger() Logger {
	return &zapLogger{util: zap.NewNop()}
}

// gcpLevelEncoder แปลง Level ให้ตรงกับ Google Cloud Severity
func gcpLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("DEBUG")
	case zapcore.InfoLevel:
		enc.AppendString("INFO")
	case zapcore.WarnLevel:
		enc.AppendString("WARNING")
	case zapcore.ErrorLevel:
		enc.AppendString("ERROR")
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		enc.AppendString("CRITICAL")
	default:
		enc.AppendString("DEFAULT")
	}
}

// WithContext แบบ Logger เข้าไปใน Context พร้อมกับ Field เริ่มต้น
func WithContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// Ctx ดึง Logger ออกมาจาก context ใน Layer ต่างๆ (ถ้าไม่มีจะส่ง Global Log กลับไป)
func Ctx(ctx context.Context) Logger {
	if ctx == nil {
		return nil
	}
	if l, ok := ctx.Value(ctxKey{}).(Logger); ok {
		return l
	}
	return nil
}
