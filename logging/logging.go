package logging

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/sqc157400661/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

// 只能输出结构化日志，但是性能要高于 SugaredLogger
var Logger *zap.Logger

// 可以输出 结构化日志、非结构化日志。性能差于 zap.Logger
var SugarLogger *zap.SugaredLogger

type LoggerOpt struct {
	LogLevel    string
	StderrLevel string
	LogPath     string
}

func Setup(opt LoggerOpt) {
	var err error
	// 从配置中获取输出到文件的日志等级
	var logLevel zapcore.Level
	logLevel, err = zapcore.ParseLevel(opt.LogLevel)
	if err != nil {
		util.PrintFatalError(err)
	}
	//  从配置中获取输出控制台的日志等级
	var stdoutLevel zapcore.Level
	stdoutLevel, err = zapcore.ParseLevel(opt.StderrLevel)
	if err != nil {
		util.PrintFatalError(err)
	}
	Logger = NewZapLogger(logLevel, stdoutLevel, "application", opt.LogPath)
	SugarLogger = Logger.Sugar()
}

func NewZapLogger(logLevel, stdoutLevel zapcore.Level, prefix, logPath string) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	//自定义日志级别：自定义Info级别
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.WarnLevel && lvl >= logLevel
	})
	//自定义日志级别：自定义Warn级别
	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel && lvl >= logLevel
	})
	// 实现多个输出
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), GetlogWriter(logPath, prefix, "info"), infoLevel),                        //将info及以下写入logPath
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), GetlogWriter(logPath, prefix, "error"), warnLevel),                       //warn及以上
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)), stdoutLevel), //同时将日志输出到控制台
	)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.WarnLevel))
}

func GetlogWriter(logPath string, prefix, tag string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logPath, fmt.Sprintf("%s_%s.log", prefix, tag)), // 文件位置
		MaxSize:    2048,                                                          // 进行切割之前,日志文件的最大大小(MB为单位)
		MaxAge:     30,                                                            // 保留旧文件的最大天数
		MaxBackups: 30,                                                            // 保留旧文件的最大个数
		Compress:   false,                                                         // 是否压缩/归档旧文件
	}
	return zapcore.AddSync(lumberJackLogger)
}
