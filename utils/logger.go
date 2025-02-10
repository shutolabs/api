package utils

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var sugar *zap.SugaredLogger

func init() {
	sugar = zap.NewNop().Sugar()
}

func InitLogger(level string) error {
	zapLevel, err := parseLogLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level %q: %w", level, err)
	}

	isDevelopment := os.Getenv("APP_ENV") == "development"
	config := createLoggerConfig(zapLevel, isDevelopment)

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return fmt.Errorf("failed to build logger: %w", err)
	}

	sugar = logger.Sugar()
	return nil
}

func parseLogLevel(level string) (zapcore.Level, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return zapLevel, err
	}
	return zapLevel, nil
}

func createLoggerConfig(level zapcore.Level, isDevelopment bool) zap.Config {
	var encoderConfig zapcore.EncoderConfig
	var encoder string

	if isDevelopment {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = "console"
	} else {
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = "json"
	}

	return zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       isDevelopment,
		Encoding:         encoder,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		DisableStacktrace: !isDevelopment,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
	}
}

// Debug logs a message at debug level with structured context
func Debug(msg string, fields ...interface{}) {
	sugar.Debugw(msg, fields...)
}

// Info logs a message at info level with structured context
func Info(msg string, fields ...interface{}) {
	sugar.Infow(msg, fields...)
}

// Warn logs a message at warn level with structured context
func Warn(msg string, fields ...interface{}) {
	sugar.Warnw(msg, fields...)
}

// Error logs a message at error level with structured context
func Error(msg string, fields ...interface{}) {
	sugar.Errorw(msg, fields...)
}

// Fatal logs a message at fatal level with structured context and then exits
func Fatal(msg string, fields ...interface{}) {
	sugar.Fatalw(msg, fields...)
}
