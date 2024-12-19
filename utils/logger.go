package utils

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sugar *zap.SugaredLogger

func init() {
	// Initialize with a no-op logger by default
	nop := zap.NewNop().Sugar()
	sugar = nop
}

func InitLogger(level string) error {
	// Convert string level to zapcore.Level
	var zapLevel zapcore.Level
	err := zapLevel.UnmarshalText([]byte(level))
	if err != nil {
		return err
	}

	isDevelopment := os.Getenv("APP_ENV") == "development"

	// Configure encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	
	encoder := "json"
	if isDevelopment {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoder = "console"
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       isDevelopment,
		Encoding:         encoder,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		DisableStacktrace: true,  // Disable stacktrace globally
	}

	logger, err := config.Build(
		zap.AddCallerSkip(1),    // Skip the logger wrapper
	)
	if err != nil {
		return err
	}

	sugar = logger.Sugar()
	return nil
}

func Debug(msg string, fields ...interface{}) {
	sugar.Debugw(msg, fields...)
}

func Info(msg string, fields ...interface{}) {
	sugar.Infow(msg, fields...)
}

func Warn(msg string, fields ...interface{}) {
	sugar.Warnw(msg, fields...)
}

func Error(msg string, fields ...interface{}) {
	sugar.Errorw(msg, fields...)
}

func Fatal(msg string, fields ...interface{}) {
	sugar.Fatalw(msg, fields...)
}
