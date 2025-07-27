package logger

import (
	"context"
	"github.com/novel/internal/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"strings"
)

var (
	log         *zap.Logger
	sugar       *zap.SugaredLogger
	atomicLevel zap.AtomicLevel
)

type ctxKeyLogger struct{}

// InitLogger 初始化日志系统，支持 console / file，支持动态日志等级
func InitLogger(cfg *config.Config) error {
	logCfg := cfg.Logger

	// 设置日志等级
	atomicLevel = zap.NewAtomicLevel()
	atomicLevel.SetLevel(parseLogLevel(logCfg.Level))

	// 编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	var encoder zapcore.Encoder
	var writer zapcore.WriteSyncer

	// 根据环境区分输出方式
	if logCfg.Mode == "prod" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)

		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   filepath.Clean(logCfg.FilePath),
			MaxSize:    logCfg.MaxSize,
			MaxBackups: logCfg.MaxBackups,
			MaxAge:     logCfg.MaxAge,
			Compress:   logCfg.Compress,
		})

		stdoutWriter := zapcore.AddSync(os.Stdout)

		// 组合多个输出目标：文件 + 标准输出
		writer = zapcore.NewMultiWriteSyncer(fileWriter, stdoutWriter)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
		writer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writer, atomicLevel)
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = log.Sugar()

	zap.ReplaceGlobals(log)

	return nil
}

func parseLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// SetLevel 动态设置日志等级
func SetLevel(level string) {
	atomicLevel.SetLevel(parseLogLevel(level))
	log.Info("日志等级已更新", zap.String("new_level", level))
}

// Sync 写入所有缓冲日志，通常在程序退出时调用
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

// ---------------- context-aware 日志封装 ----------------

func InjectLoggerIntoContext(ctx context.Context, fields ...zap.Field) context.Context {
	return context.WithValue(ctx, ctxKeyLogger{}, log.With(fields...))
}

func getLoggerFromContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return log
	}
	if loggerFromCtx, ok := ctx.Value(ctxKeyLogger{}).(*zap.Logger); ok {
		return loggerFromCtx
	}
	return log
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	getLoggerFromContext(ctx).Debug(msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	getLoggerFromContext(ctx).Info(msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	getLoggerFromContext(ctx).Warn(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	getLoggerFromContext(ctx).Error(msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	getLoggerFromContext(ctx).Fatal(msg, fields...)
}

// ---------------- 非 context 日志封装 ----------------

func DebugRaw(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

func InfoRaw(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

func WarnRaw(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

func ErrorRaw(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

func FatalRaw(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

// ---------------- 格式化日志封装 ----------------

func Debugf(format string, args ...interface{}) {
	sugar.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	sugar.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	sugar.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	sugar.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	sugar.Fatalf(format, args...)
}
