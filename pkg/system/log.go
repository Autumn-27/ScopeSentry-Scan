// Package logMode -----------------------------
// @file      : log.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2023/12/9 20:07
// -------------------------------------------
package system

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/redisClient"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

type CustomLog struct {
	Status string
	Msg    string
}

var redisLogClientInstance *redisClient.RedisClient

func PrintLog(l CustomLog) {
	timeNow := GetTimeNow()
	fmt.Printf("%s - [%s] %s\n", timeNow, l.Status, l.Msg)
	if redisLogClientInstance == nil || redisLogClientInstance.Client() == nil {
		redisAddr := AppConfig.Redis.IP + ":" + AppConfig.Redis.Port
		redisPassword := AppConfig.Redis.Password
		// 如果 redisClientInstance 为空或连接已断开，则创建新的 Redis 连接
		newClient, err := redisClient.NewRedisClient(redisAddr, redisPassword, 0)
		if err != nil {
			fmt.Printf("Failed to connect to Redis: %v\n", err)
			// 可以记录到本地日志文件或者输出到控制台
			return
		}
		redisLogClientInstance = newClient
	}

	ctx := context.Background()
	err := sendLog(ctx, fmt.Sprintf("%s - [%s] %s\n", timeNow, l.Status, l.Msg))
	if err != nil {
		fmt.Printf("Failed to send log to Redis: %v\n", err)
	}
}

type logMessage struct {
	Name string `json:"name"`
	Log  string `json:"log"`
}

func sendLog(ctx context.Context, logData string) error {
	logMsg := logMessage{
		Name: AppConfig.System.NodeName,
		Log:  logData,
	}
	msgJSON, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal log message to JSON: %v", err)
	}
	err = redisLogClientInstance.Publish(ctx, "logs", msgJSON)
	if err != nil {
		fmt.Errorf("failed to send log message : %v", err)
	}
	return nil
}

func Debug(msg string) {
	if DebugFlag {
		fmt.Printf(msg)
	}
}

var ZapLog *zap.Logger

func LogInit(flag bool) {
	logTmFmtWithMS := "2006-01-02 15:04:05"
	// 自定义时间输出格式
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(logTmFmtWithMS) + "]")
	}

	encoderConfig := zapcore.EncoderConfig{
		CallerKey:      "caller_line", // 打印文件名和行数
		LevelKey:       "level_name",
		MessageKey:     "msg",
		TimeKey:        "ts",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeTime:     customTimeEncoder, // 自定义时间格式
		NameKey:        "logger",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
	}
	atom := zap.NewAtomicLevelAt(zap.InfoLevel)
	// 设置日志级别
	if flag {
		atom = zap.NewAtomicLevelAt(zap.DebugLevel)
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}
	config := zap.Config{
		Level:         atom,               // 日志级别
		Encoding:      "console",          // 输出格式 console 或 json
		EncoderConfig: encoderConfig,      // 编码器配置
		OutputPaths:   []string{"stdout"}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
	}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder //这里可以指定颜色
	// 构建日志
	var err error
	ZapLog, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}

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
	if redisLogClientInstance == nil || redisLogClientInstance.Client() == nil {
		redisAddr := AppConfig.Redis.IP + ":" + AppConfig.Redis.Port
		redisPassword := AppConfig.Redis.Password
		// 如果 redisClientInstance 为空或连接已断开，则创建新的 Redis 连接
		newClient, err := redisClient.NewRedisClient(redisAddr, redisPassword, 0)
		if err != nil {
			fmt.Printf("Failed to connect to Redis: %v\n", err)
			// 可以记录到本地日志文件或者输出到控制台
		}
		redisLogClientInstance = newClient
	}
	ctx := context.Background()
	logMsg := logMessage{
		Name: AppConfig.System.NodeName,
		Log:  msg,
	}
	msgJSON, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal log message to JSON: %v", err)
	}
	err = redisLogClientInstance.Publish(ctx, "logs", msgJSON)
	if err != nil {
		fmt.Errorf("failed to send log message : %v", err)
	}
	return nil
}
