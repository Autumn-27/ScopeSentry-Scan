// logger-------------------------------------
// @file      : logger.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/6 22:40
// -------------------------------------------

package logger

// ZapLog Global variable to hold the logger instance
//var ZapLog *zap.Logger
//
//func LogInit(flag bool) {
//	logTmFmtWithMS := "2006-01-02 15:04:05"
//	// 自定义时间输出格式
//	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
//		enc.AppendString("[" + t.Format(logTmFmtWithMS) + "]")
//	}
//
//	encoderConfig := zapcore.EncoderConfig{
//		CallerKey:      "caller_line", // 打印文件名和行数
//		LevelKey:       "level_name",
//		MessageKey:     "msg",
//		TimeKey:        "ts",
//		LineEnding:     zapcore.DefaultLineEnding,
//		EncodeDuration: zapcore.SecondsDurationEncoder,
//		EncodeTime:     customTimeEncoder, // 自定义时间格式
//		NameKey:        "logger",
//		FunctionKey:    zapcore.OmitKey,
//		StacktraceKey:  "stacktrace",
//	}
//	atom := zap.NewAtomicLevelAt(zap.InfoLevel)
//	// 设置日志级别
//	if flag {
//		atom = zap.NewAtomicLevelAt(zap.DebugLevel)
//		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
//	}
//	config := zap.Config{
//		Level:         atom,               // 日志级别
//		Encoding:      "console",          // 输出格式 console 或 json
//		EncoderConfig: encoderConfig,      // 编码器配置
//		OutputPaths:   []string{"stdout"}, // 输出到标准输出
//	}
//	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // 颜色编码
//	// 构建日志
//	var err error
//	ZapLog, err = config.Build()
//	if err != nil {
//		panic(fmt.Sprintf("log 初始化失败: %v", err))
//	}
//}
//
//func SlogInfo(msg string) {
//	ZapLog.WithOptions(zap.AddCallerSkip(1)).Info(msg)
//	timeNow := GetTimeNow()
//	err := SendLogToRedis(fmt.Sprintf("%s - [%s] %s\n", timeNow, "INFO", msg))
//	if err != nil {
//		ZapLog.Error("Failed to send log to Redis", zap.Error(err))
//	}
//}
//
//func SlogInfoLocal(msg string) {
//	ZapLog.WithOptions(zap.AddCallerSkip(1)).Info(msg)
//}
//
//func SlogError(msg string) {
//	ZapLog.WithOptions(zap.AddCallerSkip(1)).Error(msg)
//	timeNow := GetTimeNow()
//	err := SendLogToRedis(fmt.Sprintf("%s - [%s] %s\n", timeNow, "ERROR", msg))
//	if err != nil {
//		ZapLog.Error("Failed to send log to Redis", zap.Error(err))
//	}
//}
//
//func SlogErrorLocal(msg string) {
//	ZapLog.WithOptions(zap.AddCallerSkip(1)).Error(msg)
//}
//
//func SlogDebug(msg string) {
//	ZapLog.WithOptions(zap.AddCallerSkip(1)).Debug(msg)
//	timeNow := GetTimeNow()
//	err := SendLogToRedis(fmt.Sprintf("%s - [%s] %s\n", timeNow, "DEBUG", msg))
//	if err != nil {
//		ZapLog.Error("Failed to send log to Redis", zap.Error(err))
//	}
//}
//
//func SlogDebugLocal(msg string) {
//	ZapLog.WithOptions(zap.AddCallerSkip(1)).Debug(msg)
//}
//
//func SendLogToRedis(msg string) error {
//	if redisLogClientInstance == nil || redisLogClientInstance.Client() == nil {
//		redisAddr := config.AppConfig.Redis.IP + ":" + config.AppConfig.Redis.Port
//		redisPassword := config.AppConfig.Redis.Password
//		// 如果 redisClientInstance 为空或连接已断开，则创建新的 Redis 连接
//		newClient, err := redisClient.NewRedisClient(redisAddr, redisPassword, 0)
//		if err != nil {
//			ZapLog.Error("Failed to connect to Redis", zap.Error(err))
//			return err
//		}
//		redisLogClientInstance = newClient
//	}
//	ctx := context.Background()
//	logMsg := logMessage{
//		Name: config.AppConfig.NodeName,
//		Log:  msg,
//	}
//	msgJSON, err := json.Marshal(logMsg)
//	if err != nil {
//		return fmt.Errorf("failed to marshal log message to JSON: %v", err)
//	}
//	err = redisLogClientInstance.Publish(ctx, "logs", msgJSON)
//	if err != nil {
//		return fmt.Errorf("failed to send log message: %v", err)
//	}
//	return nil
//}
//
//// GetTimeNow 获取当前时间字符串
//func GetTimeNow() string {
//	return time.Now().Format("2006-01-02 15:04:05")
//}
//
//// logMessage 是发送到 Redis 的日志消息结构体
//type logMessage struct {
//	Name string `json:"name"`
//	Log  string `json:"log"`
//}
