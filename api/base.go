package base

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sqc157400661/helper/api/common"
	"github.com/sqc157400661/helper/api/response"
	"go.uber.org/zap"
	"net/http"
)

const UserKey = "user"
const IsStaff = "isStaff"

type RequestIface interface {
	SetEngine(engine string)
}

type Api struct {
	Context *gin.Context
	Logger  *zap.SugaredLogger
	Return  *response.Return
	Errors  error
}

func (e *Api) AddError(err error) {
	if e.Errors == nil {
		e.Errors = err
	} else if err != nil {
		e.Logger.Error(err)
		e.Errors = fmt.Errorf("%v; %w", e.Errors, err)
	}
}

func (e *Api) SetResponse(responses response.Responses) {
	if responses != nil {
		e.Return = response.NewReturn(responses)
	}
}

// MakeContext 设置http上下文
func (e *Api) MakeContext(c *gin.Context) *Api {
	e.Context = c
	e.Logger = common.Logger(c)
	if e.Return == nil {
		e.Return = response.DefaultReturn
	}
	return e
}

func (e *Api) Bind(d interface{}, bindings ...binding.Binding) *Api {
	var err error
	if len(bindings) == 0 {
		if e.Context.Request.Method == http.MethodGet {
			err = e.Context.ShouldBindQuery(d)
		} else {
			err = e.Context.ShouldBind(d)
		}
		if err != nil && err.Error() == "EOF" {
			e.Logger.Warn("request body is not present anymore. ")
			err = nil
		}
		if err != nil {
			e.AddError(err)
		}
	} else {
		for i := range bindings {
			if bindings[i] == nil {
				err = e.Context.ShouldBindUri(d)
			} else {
				err = e.Context.ShouldBindWith(d, bindings[i])
			}
			if err != nil && err.Error() == "EOF" {
				e.Logger.Warn("request body is not present anymore. ")
				err = nil
				continue
			}
			if err != nil {
				e.AddError(err)
				break
			}
		}
	}
	return e
}

// Bind 参数校验和数据绑定 可参考https://gin-gonic.com/zh-cn/docs/examples/binding-and-validation/
func (e *Api) BindWithEngine(d RequestIface, bindings ...binding.Binding) *Api {
	e.Bind(d, bindings...)
	d.SetEngine(e.GetEngine())
	return e
}

func (e *Api) GetEngine() string {
	if e.Context.Request.Header.Get("engine") == "" {
		return e.Context.Request.Header.Get("Engine")
	}
	return e.Context.Request.Header.Get("engine") //请求头部
}

// Error 通常错误数据处理
func (e Api) Error(code int, err error, msg string) {
	e.Return.Error(e.Context, code, err, msg)
}

// OK 通常成功数据处理
func (e Api) OK(data interface{}, msg string) {
	e.Return.OK(e.Context, data, msg)
}

// PageOK 分页数据处理
func (e Api) PageOK(result interface{}, count int, pageIndex int, pageSize int, msg string) {
	e.Return.PageOK(e.Context, result, count, pageIndex, pageSize, msg)
}

// Custom 兼容函数
func (e Api) Custom(data gin.H) {
	e.Return.Custum(e.Context, data)
}
