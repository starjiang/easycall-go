package easycall

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/starjiang/elog"
)

//ServiceInfo for EasyServiceContext
type ServiceInfo struct {
	name    string
	port    int
	weight  int
	service interface{}
}

type MiddlewareFunc func(req *Request, resp *Response, client *EasyConnection, next *MiddlewareInfo)

type MiddlewareInfo struct {
	Middleware MiddlewareFunc
	Next       *MiddlewareInfo
}

//EasyServiceContext for Easycall
type EasyServiceContext struct {
	serviceList map[string]*ServiceInfo
	endpoints   []string
	middlewares map[string][]*MiddlewareInfo
}

func NewEasyServiceContext(endpoints []string) *EasyServiceContext {

	return &EasyServiceContext{make(map[string]*ServiceInfo, 0), endpoints, make(map[string][]*MiddlewareInfo, 0)}
}

func (svc *EasyServiceContext) CreateService(name string, port int, service interface{}, weight int) error {
	info := &ServiceInfo{name, port, weight, service}
	svc.serviceList[name] = info
	return nil
}

func (svc *EasyServiceContext) AddMiddleware(name string, middleware MiddlewareFunc) {
	list := svc.middlewares[name]
	if list == nil {
		list = make([]*MiddlewareInfo, 0)
	}
	minfo := &MiddlewareInfo{middleware, nil}

	mlen := len(list)
	if mlen > 0 {
		list[mlen-1].Next = minfo
	}

	list = append(list, minfo)
	svc.middlewares[name] = list
}

func (svc *EasyServiceContext) StartAndWait() error {

	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				for _, info := range svc.serviceList {
					elog.Infof("unregister service %s,port=%d", info.name, info.port)
					register, err := NewServiceRegister(svc.endpoints, time.Second*ETCD_CONNECT_TIMEOUT)
					if err != nil {
						continue
					}
					register.Unregister(info.name, info.port)
				}
				os.Exit(0)
			default:
				elog.Info("other signal:", s)
			}
		}
	}()

	var wg sync.WaitGroup
	size := len(svc.serviceList)
	wg.Add(size)
	for _, info := range svc.serviceList {
		server := &Server{}
		go func(info *ServiceInfo, wg *sync.WaitGroup) {
			register, err := NewServiceRegister(svc.endpoints, time.Second*ETCD_CONNECT_TIMEOUT)
			if err != nil {
				wg.Done()
				elog.Error("register fail:", err, info.name, info.port, info.weight)
				return
			}
			err = register.Register(info.name, info.port, info.weight)
			if err != nil {
				wg.Done()
				elog.Error("register fail:", err, info.name, info.port, info.weight)
				return
			}
			elog.Infof("service %s start at port %d", info.name, info.port)
			err = server.CreateServer(info.port, info.service, svc.middlewares[info.name])
			if err != nil {
				elog.Error("start service fail:", err, info.name, info.port, info.weight)
				register.Unregister(info.name, info.port)
				wg.Done()
			}
		}(info, &wg)
	}

	wg.Wait()

	return nil
}
