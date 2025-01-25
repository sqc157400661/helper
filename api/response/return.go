package response

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

const TrafficKey = "X-Request-Id"

type Return struct {
	resp Responses
}

var DefaultReturn = &Return{resp: &defaultResponse{}}

func NewReturn(resp Responses) *Return {
	return &Return{
		resp: resp,
	}
}

// Error 失败数据处理
func (r *Return) Error(c *gin.Context, code int, err error, msg string) {
	res := r.resp.Clone()
	if err != nil {
		msg = msg + " " + err.Error()
	}
	if msg != "" {
		res.SetMsg(msg)
	}
	res.SetTraceID(GenerateMsgIDFromContext(c))
	res.SetCode(int32(code))
	res.SetSuccess(false)
	c.Set("result", res)
	c.Set("status", code)
	c.AbortWithStatusJSON(http.StatusOK, res)
}

// Massage 返回massage
func (r *Return) Massage(c *gin.Context, code int, msg string) {
	res := r.resp.Clone()
	res.SetSuccess(true)
	if msg != "" {
		res.SetMsg(msg)
	}
	res.SetTraceID(GenerateMsgIDFromContext(c))
	res.SetCode(int32(code))
	c.Set("result", res)
	c.Set("status", http.StatusOK)
	c.AbortWithStatusJSON(http.StatusOK, res)
}

// OK 通常成功数据处理
func (r *Return) OK(c *gin.Context, data interface{}, msg string) {
	res := r.resp.Clone()
	res.SetData(data)
	res.SetSuccess(true)
	if msg != "" {
		res.SetMsg(msg)
	}
	res.SetTraceID(GenerateMsgIDFromContext(c))
	res.SetCode(http.StatusOK)
	c.Set("result", res)
	c.Set("status", http.StatusOK)
	c.AbortWithStatusJSON(http.StatusOK, res)
}

// PageOK 分页数据处理
func (r *Return) PageOK(c *gin.Context, result interface{}, count int, pageIndex int, pageSize int, msg string) {
	var res page
	res.List = result
	res.Count = count
	res.PageIndex = pageIndex
	res.PageSize = pageSize
	r.OK(c, res, msg)
}

// Custum 兼容函数
func (r *Return) Custum(c *gin.Context, data gin.H) {
	data["requestId"] = GenerateMsgIDFromContext(c)
	c.Set("result", data)
	c.AbortWithStatusJSON(http.StatusOK, data)
}

// GenerateMsgIDFromContext 生成msgID
func GenerateMsgIDFromContext(c *gin.Context) string {
	requestId := c.GetHeader(TrafficKey)
	if requestId == "" {
		requestId = uuid.New().String()
		c.Header(TrafficKey, requestId)
	}
	return requestId
}
