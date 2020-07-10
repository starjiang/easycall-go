package easycall

import (
	"reflect"
	"time"

	"github.com/starjiang/elog"
)

//ServiceHandler for EasyService
type ServiceHandler struct {
	service     interface{}
	value       reflect.Value
	middlewares []*MiddlewareInfo
}

func NewServiceHandler(service interface{}, middlewares []*MiddlewareInfo) *ServiceHandler {
	serviceHandler := &ServiceHandler{}
	serviceHandler.service = service
	serviceHandler.value = reflect.ValueOf(service)
	mlen := len(middlewares)

	finalFunc := func(req *Request, resp *Response, client *EasyConnection, next *MiddlewareInfo) {
		serviceHandler.onRequest(req, resp, client)
	}

	final := &MiddlewareInfo{finalFunc, nil}

	if mlen > 0 {
		middlewares[mlen-1].Next = final
	} else {
		middlewares = append(middlewares, final)
	}
	serviceHandler.middlewares = middlewares
	return serviceHandler
}

func (h *ServiceHandler) Dispatch(pkgData []byte, client *EasyConnection) {

	reqPkg, err := DecodeWithBodyData(pkgData)

	if err != nil {
		elog.Error("decode pkg fail:", err)
		return
	}

	req := &Request{reqPkg.GetFormat(), reqPkg.GetHead(), reqPkg.GetBodyData(), time.Now(), client.conn.RemoteAddr().String()}
	resp := &Response{reqPkg.GetFormat(), nil, nil}

	h.middlewares[0].Middleware(req, resp, client, h.middlewares[0].Next)
}

func (h *ServiceHandler) onRequest(req *Request, resp *Response, client *EasyConnection) {

	defer PanicHandler()
	m := h.value.MethodByName(req.head.Method)

	if !m.IsValid() {
		req.head.SetRet(ERROR_METHOD_NOT_FOUND)
		req.head.SetMsg("method " + req.head.Method + " not found")
		respPkg := NewPackageWithBody(req.format, req.head, make(map[string]interface{}))
		pkgData, err := respPkg.EncodeWithBody()
		if err != nil {
			elog.Error("encode pkg fail:", err)
			return
		}
		client.Send(pkgData)
		return
	}

	in := []reflect.Value{
		reflect.ValueOf(req),
		reflect.ValueOf(resp),
	}

	m.Call(in)

	respPkg := NewPackageWithBody(resp.format, resp.head, resp.body)

	pkgData, err := respPkg.EncodeWithBody()
	if err != nil {
		elog.Error("encode pkg fail:", err)
		return
	}
	client.Send(pkgData)
}