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
	SendPluginLogToRedis(key, fmt.Sprintf("[%v] [%v] %v", global.AppConfig.NodeName, GetTimeNow(), msg))
}

func SendPluginLogToRedis(key string, msg string) {
	ctx := context.Background()
	_, err := redis.RedisClient.SAdd(ctx, key, msg)
	if err != nil {
		SlogError(fmt.Sprintf("SendPluginLogToRedis sadd error %v", err))
	}
}

var timeZoneOffsets = map[string]int{
	"UTC":                 0,
	"Asia/Shanghai":       8 * 60 * 60,
	"Asia/Tokyo":          9 * 60 * 60,
	"Asia/Kolkata":        5*60*60 + 30*60,
	"Europe/London":       0,
	"Europe/Berlin":       1 * 60 * 60,
	"Europe/Paris":        1 * 60 * 60,
	"America/New_York":    -5 * 60 * 60,
	"America/Chicago":     -6 * 60 * 60,
	"America/Denver":      -7 * 60 * 60,
	"America/Los_Angeles": -8 * 60 * 60,
	"Australia/Sydney":    10 * 60 * 60,
	"Australia/Perth":     8 * 60 * 60,
	"Asia/Singapore":      8 * 60 * 60,
	"Asia/Hong_Kong":      8 * 60 * 60,
	"Europe/Moscow":       3 * 60 * 60,
	"America/Sao_Paulo":   -3 * 60 * 60,
	"Africa/Johannesburg": 2 * 60 * 60,
	"Asia/Dubai":          4 * 60 * 60,
	"Pacific/Auckland":    12 * 60 * 60,
}

func GetTimeNow() string {
	// 获取当前时间
	timeZoneName := global.AppConfig.TimeZoneName

	var location *time.Location
	var err error

	// 查找时区名称对应的偏移量
	offset, exists := timeZoneOffsets[timeZoneName]
	if exists {
		// 如果存在映射，使用固定时区
		location = time.FixedZone(timeZoneName, offset)
	} else {
		// 如果映射不存在，尝试直接加载时区名称
		location, err = time.LoadLocation(timeZoneName)
		if err != nil {
			// 如果加载失败，使用系统默认时区
			fmt.Printf("Time zone not found: %s, using system default time zone\n", timeZoneName)
			location = time.Local
		}
	}
	currentTime := time.Now()
	var easternTime = currentTime.In(location)
	return easternTime.Format("2006-01-02 15:04:05")
}
