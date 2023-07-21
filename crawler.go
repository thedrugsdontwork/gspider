package GSpider

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type _signal int

const (
	STOP    _signal = 0
	RUNNING _signal = 1
	WAITING _signal = 2
)

type crawler interface {
	/*
		初始化线程池
	*/
	Init(poolSize int)
	/*
		用于获取线程池大小
	*/
	GetPoolSize() int
	/*
		注册请求前置处理中间件
	*/
	RegisterRequestMiddleware(middleWare *RequestMiddleware)
	/*
		注册响应前置处理中间件
	*/
	RegisterResponseMiddleware(middleWare *ResponseMiddleware)
	/*
		注册采集起始节点
	*/
	RegisterStartingRequest(f *func() []*BaseRequestObj)
	/*
		开始采集
	*/
	StartCrawl()
	/*
		停止采集
	*/
	ShutDown()
}

type Crawler struct {
	poolSize            int
	crawlerName         string
	requestMiddlewares  []*RequestMiddleware
	responseMiddlewares []*ResponseMiddleware
	reqMSize            int
	resMSize            int
	startFunc           *func() []*BaseRequestObj
	requests            chan *BaseRequestObj
	requestsCurSize     int32
	responses           chan *BaseResponseObj
	responseCurSize     int32
	signals             []_signal
	rCounter            int32
	reqIngCounter       int32
	parIngCounter       int32
	wg                  sync.WaitGroup
	signal              _signal
}

func (c *Crawler) Init(crawlerName string, poolSize int) {
	c.poolSize = 4
	c.crawlerName = crawlerName
	c.requests = make(chan *BaseRequestObj, 100)
	c.responses = make(chan *BaseResponseObj, 100)

}
func (c *Crawler) GetPoolSize() int {
	return c.poolSize
}
func (c *Crawler) RegisterRequestMiddleware(middleWare *RequestMiddleware) {
	c.requestMiddlewares = append(c.requestMiddlewares, middleWare)
	c.reqMSize++
}

func (c *Crawler) RegisterResponseMiddleware(middleWare *ResponseMiddleware) {
	c.responseMiddlewares = append(c.responseMiddlewares, middleWare)
	c.resMSize++
}

func (c *Crawler) RegisterStartingRequest(f *(func() []*BaseRequestObj)) {
	if c.startFunc != nil {
		panic("Start function already registered.")
	}
	c.startFunc = f
}

/*
	首先需要运行start func 获取第一批request obj，
	待请求发送后
*/
func (c *Crawler) StartCrawl() {
	fmt.Printf("Crawler %s has been running", c.crawlerName)
	c.wg.Add(12)
	var counter uint32 = 0
	var reqs []*BaseRequestObj = (*c.startFunc)()
	c.download()
	submitRequest := func(reqs []*BaseRequestObj, cr *uint32) {
		for _, req := range reqs {
			if req != nil {
				fmt.Printf("Add the request:%s\n", req.URL)
				atomic.AddInt32(&c.requestsCurSize, 1)
				fmt.Printf("Add int32:%s\n", req.URL)
				c.requests <- req
				fmt.Printf("Add the request:%s\n", req.URL)

				fmt.Printf("C++t:%s\n", req.URL)
			}

		}
	}
	c.signal = WAITING
	go func() {
		submitRequest(reqs, &counter)
		c.signal = RUNNING
		c.wg.Done()

	}()
	if reqs == nil {
		panic("The start function return nil except slice of BaseResponseObj pointer")
	}

	fmt.Printf("Start function total submit reuqest %d", counter)
	/*do downloading*/

	/*do parsing*/

	for i := 0; i < 10; i++ {
		go func() {

			for res := range c.responses {
				atomic.AddInt32(&c.responseCurSize, -1)
				fmt.Printf("Start parse page %d", res.Request.RetryTimes)
				atomic.AddInt32(&c.parIngCounter, 1)
				if res.StatusCode < 400 {
					if res.Request.Callback != nil {
						reqs = res.Request.Callback(res)
					}
				} else {
					if res.Request.Errback != nil {
						reqs = res.Request.Errback(res)
					}
				}
				if reqs != nil {
					submitRequest(reqs, &counter)
				}
				atomic.AddInt32(&c.parIngCounter, -1)
			}
			c.wg.Done()
		}()
	}
	/*wait until stop && check if complete*/
	go func() {
		for c.signal != RUNNING || atomic.LoadInt32(&c.reqIngCounter) != 0 || atomic.LoadInt32(&c.parIngCounter) != 0 || atomic.LoadInt32(&c.requestsCurSize) != 0 || atomic.LoadInt32(&c.responseCurSize) != 0 {
			fmt.Println(atomic.LoadInt32(&c.reqIngCounter), atomic.LoadInt32(&c.parIngCounter), atomic.LoadInt32(&c.requestsCurSize), atomic.LoadInt32(&c.responseCurSize))
			time.Sleep(1 * time.Second)
		}

		c.ShutDown()
		c.wg.Done()
	}()
	/*statistic all target*/
	//go func() {
	//
	//	for {
	//		time.Sleep(60 * time.Second)
	//		fmt.Printf("The download status is ...")
	//	}
	//	c.wg.Done()
	//}()
	c.wg.Wait()
	fmt.Printf("Program shutdown.")
}

func (c *Crawler) ShutDown() {
	/*
		destory all var and exit
	*/
	fmt.Printf("All page has been download %d\n", c.rCounter)
	fmt.Printf("Start close the channel requests %d\n", c.requestsCurSize)
	close(c.requests)
	fmt.Printf("Start close the channel response %d\n", c.responseCurSize)
	close(c.responses)
	fmt.Printf("Crawler %s has been closed", c.crawlerName)

}
