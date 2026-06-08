package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var L *zap.Logger

func Init(mode string) error {
	var cfg zap.Config
	if mode == "debug" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.OutputPaths = []string{"stdout"}

	l, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return err
	}
	L = l
	zap.ReplaceGlobals(l)
	return nil
}

func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}

func WithTraceID(traceID string) *zap.Logger {
	if L == nil {
		return zap.NewNop()
	}
	return L.With(zap.String("trace_id", traceID))
}

func Info(msg string, fields ...zap.Field) {
	if L != nil {
		L.Info(msg, fields...)
	}
}

func Error(msg string, fields ...zap.Field) {
	if L != nil {
		L.Error(msg, fields...)
	}
}

func Fatal(msg string, fields ...zap.Field) {
	if L != nil {
		L.Fatal(msg, fields...)
	} else {
		os.Exit(1)
	}
}
