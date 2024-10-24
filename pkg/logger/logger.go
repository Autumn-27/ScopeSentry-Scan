// logger-------------------------------------
// @file      : logger.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 22:40
// -------------------------------------------

package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/global"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/redis"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

type Logger struct {
	zapLogger *zap.Logger
}

var ZapLog *zap.Logger

func NewLogger() error {
	logTmFmtWithMS := "2006-01-02 15:04:05"
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(logTmFmtWithMS) + "]")
	}

	encoderConfig := zapcore.EncoderConfig{
		LevelKey:       "level_name",
		MessageKey:     "msg",
		TimeKey:        "ts",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeTime:     customTimeEncoder,
		NameKey:        "logger",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
	}
	atom := zap.NewAtomicLevelAt(zap.InfoLevel)
	if global.AppConfig.Debug {
		atom = zap.NewAtomicLevelAt(zap.DebugLevel)
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		//encoderConfig.CallerKey = "caller_line"
	}
	c := zap.Config{
		Level:         atom,
		Encoding:      "console",
		EncoderConfig: encoderConfig,
		OutputPaths:   []string{"stdout"},
	}
	c.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	var err error
	ZapLog, err = c.Build()
	if err != nil {
		return fmt.Errorf("log 初始化失败: %v", err)
	}
	return nil
}

func SlogInfo(msg string) {
	ZapLog.WithOptions(zap.AddCallerSkip(1)).Info(msg)
	timeNow := GetTimeNow()
	err := SendLogToRedis(fmt.Sprintf("%s - [%s] %s\n", timeNow, "INFO", msg))
	if err != nil {
		return
	}
}
func SlogInfoLocal(msg string) {
	ZapLog.WithOptions(zap.AddCallerSkip(1)).Info(msg)
}
func SlogWarnLocal(msg string) {
	ZapLog.WithOptions(zap.AddCallerSkip(1)).Warn(msg)
}

func SlogWarn(msg string) {
	ZapLog.WithOptions(zap.AddCallerSkip(1)).Warn(msg)
	timeNow := GetTimeNow()
	err := SendLogToRedis(fmt.Sprintf("%s - [%s] %s\n", timeNow, "WARING", msg))
	if err != nil {
		return
	}
}

func SlogError(msg string) {
	ZapLog.WithOptions(zap.AddCallerSkip(1)).Error(msg)
	timeNow := GetTimeNow()
	err := SendLogToRedis(fmt.Sprintf("%s - [%s] %s\n", timeNow, "ERROR", msg))
	if err != nil {
		return
	}
}
func SlogErrorLocal(msg string) {
	ZapLog.WithOptions(zap.AddCallerSkip(1)).Error(msg)
}
func SlogDebug(msg string) {
	ZapLog.WithOptions(zap.AddCallerSkip(1)).Debug(msg)
	timeNow := GetTimeNow()
	err := SendLogToRedis(fmt.Sprintf("%s - [%s] %s\n", timeNow, "DEBUG", msg))
	if err != nil {
		return
	}
}
func SlogDebugLocal(msg string) {
	ZapLog.WithOptions(zap.AddCallerSkip(1)).Debug(msg)
}

func SendLogToRedis(msg string) error {
	ctx := context.Background()
	logMsg := logMessage{
		Name: global.AppConfig.NodeName,
		Log:  msg,
	}
	msgJSON, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal log message to JSON: %v", err)
	}
	err = redis.RedisClient.Publish(ctx, "logs", msgJSON)
	if err != nil {
		return fmt.Errorf("failed to send log message : %v", err)
	}
	return nil
}

func GetTimeNow() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

type logMessage struct {
	Name string `json:"name"`
	Log  string `json:"log"`
}

func PluginsLog(msg string, tp string, module string, id string) {
	switch tp {
	case "i":
		SlogInfoLocal(msg)
		msg = "[info] " + msg
	case "e":
		SlogErrorLocal(msg)
		msg = "[error] " + msg
	case "d":
		SlogDebugLocal(msg)
		msg = "[debug] " + msg
	case "w":
		SlogWarnLocal(msg)
		msg = "[warning] " + msg

	}
	key := fmt.Sprintf("logs:plugins:%v:%v", module, id)
	SendPluginLogToRedis(key, fmt.Sprintf("[%v] [%v] %v", global.AppConfig.NodeName, utils.Tools.GetTimeNow(), msg))
}

func SendPluginLogToRedis(key string, msg string) {
	ctx := context.Background()
	_, err := redis.RedisClient.SAdd(ctx, key, msg)
	if err != nil {
		SlogError(fmt.Sprintf("SendPluginLogToRedis sadd error %v", err))
	}
}
