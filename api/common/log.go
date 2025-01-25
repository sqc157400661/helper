package common

import (
	"github.com/gin-gonic/gin"
	"github.com/sqc157400661/helper/api/response"
	"github.com/sqc157400661/helper/logging"
	"go.uber.org/zap"
)

func Logger(c *gin.Context) *zap.SugaredLogger {
	if logging.SugarLogger == nil {
		return nil
	}
	return logging.SugarLogger.With(zap.String("traceid", response.GenerateMsgIDFromContext(c)))
}
