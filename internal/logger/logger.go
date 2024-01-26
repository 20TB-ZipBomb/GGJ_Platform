package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"

    "github.com/20TB-ZipBomb/GGJ_Platform/internal/utils"
)

var sugar *zap.SugaredLogger

func Init() {
    config := UseEnvConfig()
    encoderConfig := UseEnvEncoderConfig()
    encoderConfig.StacktraceKey = ""
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

func UseEnvConfig() zap.Config {
    if utils.IsProductionEnv() {
        return zap.NewProductionConfig()
    }

    return zap.NewDevelopmentConfig()
}

func UseEnvEncoderConfig() zapcore.EncoderConfig {
    if utils.IsProductionEnv() {
        return zap.NewProductionEncoderConfig()
    }

    return zap.NewDevelopmentEncoderConfig()
}

func Info(msg string, fields ...interface{}) {
    sugar.Info(msg, fields)
}

func Debug(msg string, fields ...interface{}) {
    sugar.Debug(msg, fields)
}

func Error(msg string, fields ...interface{}) {
    sugar.Error(msg, fields)
}

func Fatal(msg string, fields ...interface{}) {
    sugar.Fatal(msg, fields)
}
