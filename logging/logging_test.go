package logging

import (
	"github.com/natefinch/lumberjack"
	"log"
	"testing"
)

func TestLogger(t *testing.T) {
	logger := &lumberjack.Logger{
		Filename:   "test/myapp.log",
		MaxSize:    100, // 兆字节
		MaxBackups: 1,
		MaxAge:     28, // 天数
	}

	defer logger.Close()
	log.SetOutput(logger)

	log.Printf("%s %s %s", "instance1", "ok", "running")
}
