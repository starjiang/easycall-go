package easycall

import (
	"github.com/panjf2000/ants/v2"
	"github.com/starjiang/elog"
)

//ClientHandler for EasyService
type ClientHandler struct {
	client interface{}
	pool   *ants.Pool
}

func NewClientHandler(client interface{}) *ClientHandler {
	clientHandler := &ClientHandler{}
	clientHandler.client = client
	pool, _ := ants.NewPool(EASYCALL_CLIENT_GO_POOL_SIZE, ants.WithNonblocking(true))

	clientHandler.pool = pool
	return clientHandler
}

func (h *ClientHandler) Dispatch(pkgData []byte, client *EasyConnection) {

	err := h.pool.Submit(func() {
		defer PanicHandler()
		serviceClient := h.client.(*ServiceClient)
		reqPkg, err := DecodeWithBodyData(pkgData)
		if err != nil {
			elog.Error("decode pkg fail:", err)
		}
		serviceClient.Process(reqPkg)
	})

	if err != nil {
		elog.Error("submit to pool fail,", err)
	}
}
