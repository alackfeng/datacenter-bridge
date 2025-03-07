package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

type LogConfigure struct {
	Level       string `yaml:"level" json:"level" comment:"日志级别"`
	FilePath    string `yaml:"file_path" json:"file_path" comment:"日志文件保存路径"`
	FileMaxSize int    `yaml:"file_max_size" json:"file_max_size" comment:"日志分割前大小MB"`
	FileMaxAge  int    `yaml:"file_max_age" json:"file_max_age" comment:"日志分割前天数days"`
}

// NewLogConfigure -
func NewLogConfigure() *LogConfigure {
	return &LogConfigure{
		FilePath:    "./logs/dcb_server.log",
		FileMaxSize: 28,
		FileMaxAge:  100,
	}
}

func init() {
	logger, _ = zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
}

func logLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.DebugLevel
	}
}

// InitLogger - 初始化日志.
func InitLogger(release string, config *LogConfigure) {

	if strings.ToLower(release) == "release" {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder := zapcore.NewJSONEncoder(encoderConfig)
		var cores []zapcore.Core
		cores = append(cores,
			zapcore.NewCore(encoder, zapcore.Lock(zapcore.AddSync(&lumberjack.Logger{
				Filename: config.FilePath,
				MaxSize:  config.FileMaxAge, // megabytes
				// MaxBackups: 3,
				MaxAge:    config.FileMaxSize, // days
				LocalTime: true,
				Compress:  false,
			})), zapcore.InfoLevel),
			zapcore.NewCore(encoder, zapcore.Lock(zapcore.AddSync(os.Stdout)), logLevel(config.Level)))
		logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder := zapcore.NewConsoleEncoder(encoderConfig)
		cores := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), logLevel(config.Level))
		logger = zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	zap.ReplaceGlobals(logger)
}

func Sync() {
	logger.Sync()
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}
func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}
func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}
func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}
func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}

func Infof(msg string, args ...interface{}) {
	logger.Info(fmt.Sprintf(msg, args...))
}
func Debugf(msg string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(msg, args...))
}
func Errorf(msg string, args ...interface{}) {
	logger.Error(fmt.Sprintf(msg, args...))
}
func Warnf(msg string, args ...interface{}) {
	logger.Warn(fmt.Sprintf(msg, args...))
}
func Fatalf(msg string, args ...interface{}) {
	logger.Fatal(fmt.Sprintf(msg, args...))
}
