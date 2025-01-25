package response

type DefaultResponse struct {
	// 数据集
	TraceID string `protobuf:"bytes,1,opt,name=traceid,proto3" json:"traceid,omitempty"`
	Code    int32  `protobuf:"varint,2,opt,name=code,proto3" json:"code,omitempty"`
	Msg     string `protobuf:"bytes,3,opt,name=msg,proto3" json:"msg,omitempty"`
	Status  string `protobuf:"bytes,4,opt,name=status,proto3" json:"status,omitempty"`
}

type defaultResponse struct {
	DefaultResponse
	Data interface{} `json:"data"`
}

type Page struct {
	Count     int `json:"count"`
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
}

type page struct {
	Page
	List interface{} `json:"list"`
}

func (e *defaultResponse) SetData(data interface{}) {
	e.Data = data
}

func (e defaultResponse) Clone() Responses {
	return &e
}

func (e *defaultResponse) SetTraceID(id string) {
	e.TraceID = id
}

func (e *defaultResponse) SetMsg(s string) {
	e.Msg = s
}

func (e *defaultResponse) SetCode(code int32) {
	e.Code = code
}

func (e *defaultResponse) SetSuccess(success bool) {
	if !success {
		e.Status = "error"
	} else {
		e.Status = "success"
	}
}
