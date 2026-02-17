package log

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger(level string, outputPath string, maxSize int, maxBackups int, maxAge int, compress bool, consoleOutput bool) {
	// 配置日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   outputPath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}

	// 设置日志级别
	logLevel := zapcore.DebugLevel
	switch strings.ToLower(level) {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	// 创建zap配置
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncodeCaller = zapcore.ShortCallerEncoder

	// 创建编码器
	encoder := zapcore.NewJSONEncoder(config)

	// 创建core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(lumberjackLogger),
		logLevel,
	)

	// 如果需要同时输出到控制台，添加控制台输出
	if consoleOutput {
		consoleEncoder := zapcore.NewConsoleEncoder(config)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			logLevel,
		)
		core = zapcore.NewTee(core, consoleCore)
	}

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// Debug 输出debug级别的日志
func Debug(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Debug(msg, fields...)
	}
}

// Info 输出info级别的日志
func Info(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Info(msg, fields...)
	}
}

// Warn 输出warn级别的日志
func Warn(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Warn(msg, fields...)
	}
}

// Error 输出error级别的日志
func Error(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Error(msg, fields...)
	}
}

// Fatal 输出fatal级别的日志，并退出程序
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	} else {
		fmt.Fprintf(os.Stderr, "[FATAL] %s\n", msg)
		os.Exit(1)
	}
}

// Infof 格式化输出info级别的日志
func Infof(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Info(fmt.Sprintf(format, args...))
	}
}

// Errorf 格式化输出error级别的日志
func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Error(fmt.Sprintf(format, args...))
	}
}

// Fatalf 格式化输出fatal级别的日志，并退出程序
func Fatalf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Fatal(fmt.Sprintf(format, args...))
	} else {
		fmt.Fprintf(os.Stderr, "[FATAL] %s\n", fmt.Sprintf(format, args...))
		os.Exit(1)
	}
}