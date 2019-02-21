package service

import (
	"fmt"
	"sync"
	"time"
)

type IServiceManager interface {
	Setup(s IService) bool
	Init(logger ILogger) bool
	Start() bool
	CreateServiceID() int
}

type CServiceManager struct {
	genserviceid    int
	localserviceMap map[string]IService
	logger          ILogger
}

func (slf *CServiceManager) Setup(s IService) bool {

	slf.localserviceMap[s.GetServiceName()] = s
	return true
}

func (slf *CServiceManager) FindService(serviceName string) IService {
	service, ok := slf.localserviceMap[serviceName]
	if ok {
		return service
	}

	return nil
}

type FetchService func(s IService) error

func (slf *CServiceManager) FetchService(s FetchService) IService {
	for _, se := range slf.localserviceMap {
		s(se)
	}

	return nil
}

func (slf *CServiceManager) Init(logger ILogger, exit chan bool, pwaitGroup *sync.WaitGroup) bool {

	slf.logger = logger
	for _, s := range slf.localserviceMap {
		(s.(IModule)).InitModule(exit, pwaitGroup)
		//(s.(IModule)).OnInit()
	}

	return true
}

func (slf *CServiceManager) Start() bool {
	for _, s := range slf.localserviceMap {
		go (s.(IModule)).RunModule(s.(IModule))
	}

	return true
}

func (slf *CServiceManager) CheckServiceTimeTimeout(exit chan bool, pwaitGroup *sync.WaitGroup) {
	defer pwaitGroup.Done()
	for {
		select {
		case <-exit:
			fmt.Println("CheckServiceTimeTimeout stopping...")
			return
		}

		for _, s := range slf.localserviceMap {

			if s.IsTimeOutTick(20000) == true {
				Log.Printf("service:%s is timeout,state:%d", s.GetServiceName(), s.GetStatus())
			}
		}
		time.Sleep(2 * time.Second)
	}

}

func (slf *CServiceManager) GenServiceID() int {
	slf.genserviceid += 1
	return slf.genserviceid
}

func (slf *CServiceManager) Get() bool {
	for _, s := range slf.localserviceMap {
		go s.OnRun()
	}

	return true
}

func (slf *CServiceManager) GetLogger() ILogger {
	return slf.logger
}

var _self *CServiceManager

func InstanceServiceMgr() *CServiceManager {
	if _self == nil {
		_self = new(CServiceManager)
		_self.localserviceMap = make(map[string]IService)
		return _self
	}
	return _self
}

func GetLogger() ILogger {
	return InstanceServiceMgr().GetLogger()
}
