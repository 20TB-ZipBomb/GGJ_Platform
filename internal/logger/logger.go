package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils"
)

var sugar *zap.SugaredLogger

func Init() {
	config := useEnvConfig()
	encoderConfig := useEnvEncoderConfig()
	encoderConfig.StacktraceKey = ""
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	config.EncoderConfig = encoderConfig

	logger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic(err)
	}

	sugar = logger.Sugar()
}

func Sync() {
	sugar.Sync()
}

func Info(args ...interface{}) {
	sugar.Info(args...)
}

func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

func Debug(args ...interface{}) {
	sugar.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

func Error(args ...interface{}) {
	sugar.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

func Fatal(args ...interface{}) {
	sugar.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}

func useEnvConfig() zap.Config {
	if utils.IsProductionEnv() {
		return zap.NewProductionConfig()
	}

	return zap.NewDevelopmentConfig()
}

func useEnvEncoderConfig() zapcore.EncoderConfig {
	if utils.IsProductionEnv() {
		return zap.NewProductionEncoderConfig()
	}

	return zap.NewDevelopmentEncoderConfig()
}
