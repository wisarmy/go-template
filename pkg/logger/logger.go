// pkg/logger/logger.go
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
	once   sync.Once
)

// Config holds logger configuration
type Config struct {
	Level      string `mapstructure:"level"`
	Output     string `mapstructure:"output"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	Color      bool   `mapstructure:"color"`
}

// NewDefaultConfig returns a default logger configuration
func NewDefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Output:     "console",
		File:       "logs/app.log",
		MaxSize:    100,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
		Color:      true,
	}
}

// Init initializes the logger with configuration
func Init(cfg *Config) error {
	var err error
	once.Do(func() {
		consoleEncoderConfig := zap.NewProductionEncoderConfig()
		consoleEncoderConfig.TimeKey = "time"
		consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		if cfg.Color {
			consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		} else {
			consoleEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		}

		jsonEncoderConfig := zap.NewProductionEncoderConfig()
		jsonEncoderConfig.TimeKey = "time"
		jsonEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		jsonEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		// Determine log level
		level := getLogLevel(cfg.Level)

		// Configure cores based on output destination
		var core zapcore.Core
		switch cfg.Output {
		case "console":
			core = zapcore.NewCore(
				zapcore.NewConsoleEncoder(consoleEncoderConfig),
				zapcore.AddSync(os.Stdout),
				level,
			)
		case "file":
			// Ensure log directory exists
			logDir := filepath.Dir(cfg.File)
			if err = os.MkdirAll(logDir, 0755); err != nil {
				err = fmt.Errorf("creating log directory: %w", err)
				return
			}
			core = zapcore.NewCore(
				zapcore.NewJSONEncoder(jsonEncoderConfig),
				zapcore.AddSync(&lumberjack.Logger{
					Filename:   cfg.File,
					MaxSize:    cfg.MaxSize,
					MaxBackups: cfg.MaxBackups,
					MaxAge:     cfg.MaxAge,
					Compress:   cfg.Compress,
				}),
				level,
			)
		case "both":
			// Ensure log directory exists
			logDir := filepath.Dir(cfg.File)
			if err = os.MkdirAll(logDir, 0755); err != nil {
				err = fmt.Errorf("creating log directory: %w", err)
				return
			}
			core = zapcore.NewTee(
				zapcore.NewCore(
					zapcore.NewConsoleEncoder(consoleEncoderConfig),
					zapcore.AddSync(os.Stdout),
					level,
				),
				zapcore.NewCore(
					zapcore.NewJSONEncoder(jsonEncoderConfig),
					zapcore.AddSync(&lumberjack.Logger{
						Filename:   cfg.File,
						MaxSize:    cfg.MaxSize,
						MaxBackups: cfg.MaxBackups,
						MaxAge:     cfg.MaxAge,
						Compress:   cfg.Compress,
					}),
					level,
				),
			)
		default:
			// Default to console if invalid output type
			core = zapcore.NewCore(
				zapcore.NewConsoleEncoder(consoleEncoderConfig),
				zapcore.AddSync(os.Stdout),
				level,
			)
		}

		// Create logger
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
		sugar = logger.Sugar()
	})

	return err
}

// getLogLevel converts string log level to zapcore.Level
func getLogLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warning", "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// Debug logs a debug message
func Debug(msg string) {
	if sugar != nil {
		sugar.Debug(msg)
	}
}

// Debugf logs a debug message with formatting
func Debugf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Debugf(format, args...)
	}
}

// Info logs an info message
func Info(msg string) {
	if sugar != nil {
		sugar.Info(msg)
	}
}

// Infof logs an info message with formatting
func Infof(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Infof(format, args...)
	}
}

// Warn logs a warning message
func Warn(msg string) {
	if sugar != nil {
		sugar.Warn(msg)
	}
}

// Warnf logs a warning message with formatting
func Warnf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Warnf(format, args...)
	}
}

// Error logs an error message
func Error(msg string) {
	if sugar != nil {
		sugar.Error(msg)
	}
}

// Errorf logs an error message with formatting
func Errorf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Errorf(format, args...)
	}
}

// Fatal logs a fatal message
func Fatal(msg string) {
	if sugar != nil {
		sugar.Fatal(msg)
	}
}

// Fatalf logs a fatal message with formatting
func Fatalf(format string, args ...interface{}) {
	if sugar != nil {
		sugar.Fatalf(format, args...)
	}
}

// With returns a logger with the specified fields
func With(fields ...interface{}) *zap.SugaredLogger {
	if sugar != nil {
		return sugar.With(fields...)
	}
	return nil
}

// WithField adds a single field to the logger
func WithField(key string, value interface{}) *zap.SugaredLogger {
	if sugar != nil {
		return sugar.With(key, value)
	}
	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
