package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

type LogConfigure struct {
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

// InitLogger - 初始化日志.
func InitLogger(release bool, config *LogConfigure) {

	if release {
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
			zapcore.NewCore(encoder, zapcore.Lock(zapcore.AddSync(os.Stdout)), zapcore.InfoLevel))
		logger = zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		encoderConfig := zap.NewDevelopmentEncoderConfig()
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder := zapcore.NewConsoleEncoder(encoderConfig)
		cores := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel)
		logger = zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	zap.ReplaceGlobals(logger)
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
