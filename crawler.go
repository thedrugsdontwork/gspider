package GSpider

//todo 1.添加日志收集器
//todo 2.优化cache阈值缓存本地
//todo 3.添加下载统计（定时）
import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
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
	Init(crawlerName string, spiderSize int32, parserSize int32)
	/*
		用于获取线程池大小
	*/
	GetspiderSize() int32
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
	spiderSize          int32
	parserSize          int32
	crawlerName         string
	requestMiddlewares  []*RequestMiddleware
	responseMiddlewares []*ResponseMiddleware
	reqMSize            int
	resMSize            int
	startFunc           *func() []*BaseRequestObj
	requests            chan *BaseRequestObj
	//通道目前大小
	requestsCurSize int32
	responses       chan *BaseResponseObj
	responseCurSize int32
	signals         []_signal
	rCounter        int32
	//目前处理线程
	reqIngCounter int32
	parIngCounter int32
	wg            sync.WaitGroup
	signal        _signal
	reqCache      *BaseReqCache
	resCache      *BaseResCache
	logger        *zap.Logger
	startTime     time.Time
	ssaasd        *cache
}

func (c *Crawler) Init(crawlerName string, spiderSize int32, parserSize int32) {
	c.spiderSize = spiderSize
	c.parserSize = parserSize
	c.crawlerName = crawlerName
	c.requests = make(chan *BaseRequestObj, spiderSize)
	c.responses = make(chan *BaseResponseObj, parserSize)
	//c.ssaasd = &BaseResCache{}
	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "initialFields": {"crawlerName": "` + crawlerName + `"},
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	c.logger = zap.Must(cfg.Build())

	c.reqCache = &BaseReqCache{
		size: 0,
	}
	c.resCache = &BaseResCache{
		size: 0,
	}

}
func (c *Crawler) GetspiderSize() int32 {
	return c.spiderSize
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
func (c *Crawler) submitRequest(reqs []*BaseRequestObj) {
	//收口所有请求提交
	for _, req := range reqs {
		if req != nil {
			c.reqCache.Store(req)
			atomic.AddInt32(&c.rCounter, 1)
			atomic.AddInt32(&c.requestsCurSize, 1)
		}
	}
}
func (c *Crawler) submitResponse(res *BaseResponseObj) {
	//收口所有响应提交
	if res != nil {
		c.resCache.Store(res)
		atomic.AddInt32(&c.responseCurSize, 1)
	}
}

/*
	首先需要运行start func 获取第一批request obj，
	待请求发送后
*/
func (c *Crawler) StartCrawl() {
	defer c.logger.Sync()
	c.logger.Info("Crawler has been running.", zap.String("name", c.crawlerName))
	c.wg.Add(4)
	c.startTime = time.Now()
	c.signal = WAITING
	c.logger.Info("Start run the start function&get start reqs.")
	var reqs []*BaseRequestObj = (*c.startFunc)()
	c.logger.Info("Start run the download process.", zap.Int32("processSize", c.spiderSize))
	/*do downloading*/
	c.download()

	c.logger.Info("Start run the parse process.", zap.Int32("parseSize", c.parserSize))
	/*do parsing*/
	c.parse()

	if reqs == nil {
		c.logger.Error("Start function return nil except slice of BaseResponseObj addr.")
		panic("The start function return nil except slice of BaseResponseObj addr.")
	}
	c.submitRequest(reqs)
	c.signal = RUNNING
	c.logger.Info("Start function total submit reuqest ", zap.Int32("start request size", atomic.LoadInt32(&c.rCounter)))
	/*use to submit req to channel*/
	go func() {
		for c.signal != STOP {
			if atomic.LoadInt32(&c.reqIngCounter) >= c.spiderSize {
				continue
			}
			req := c.reqCache.Load()
			if req == nil {
				continue
			}
			c.requests <- req
		}
		c.wg.Done()
	}()
	/*use to submit response to channel*/
	go func() {
		for c.signal != STOP {
			if atomic.LoadInt32(&c.parIngCounter) >= c.parserSize {
				continue
			}
			res := c.resCache.Load()
			if res == nil {
				continue
			}
			c.responses <- res
		}
		c.wg.Done()
	}()
	/*wait until stop && check if complete*/
	go func() {
		for c.reqCache.size != 0 || c.resCache.size != 0 || c.signal != RUNNING || atomic.LoadInt32(&c.reqIngCounter) != 0 || atomic.LoadInt32(&c.parIngCounter) != 0 || atomic.LoadInt32(&c.requestsCurSize) != 0 || atomic.LoadInt32(&c.responseCurSize) != 0 {
			//fmt.Println(atomic.LoadInt32(&c.reqIngCounter), atomic.LoadInt32(&c.parIngCounter), atomic.LoadInt32(&c.requestsCurSize), atomic.LoadInt32(&c.responseCurSize))
			time.Sleep(1 * time.Second)
		}

		c.ShutDown()
		c.wg.Done()
	}()

	/*statistic all target*/
	go func() {
		defer c.logger.Sync()
		for c.signal != STOP {
			time.Sleep(1 * time.Second)
			//fmt.Printf("The download status is ...")
			now := time.Now()
			now.Sub(c.startTime)
			c.logger.Info("Crawler statistic", zap.String("name", c.crawlerName), zap.Int32("totalPage", atomic.LoadInt32(&c.rCounter)), zap.Float64("spend/s", now.Sub(c.startTime).Seconds()), zap.Int32("per3Sec/pages", atomic.LoadInt32(&c.rCounter)/int32(now.Sub(c.startTime).Seconds())))
		}
		c.wg.Done()
	}()
	c.wg.Wait()
	fmt.Printf("Program shutdown.")
}

func (c *Crawler) ShutDown() {
	/*
		destory all var and exit
	*/
	defer c.logger.Sync()
	c.logger.Info("All page has been download \n", zap.Int32("totalRequest", c.rCounter))
	c.logger.Info("Start close the channel requests \n", zap.Int32("requestsCurSize", c.requestsCurSize))
	close(c.requests)
	c.logger.Info("Start close the channel response %d\n", zap.Int32("responseCurSize", c.responseCurSize))
	close(c.responses)
	c.signal = STOP
	c.logger.Info("Crawler %s has been closed", zap.String("crawlerName", c.crawlerName))
}

func SpiderInit(c crawler, crawlerName string, spiderSize int32, parserSize int32) {
	c.Init(crawlerName, spiderSize, parserSize)
}
func GetSpiderSize(c crawler) int {
	return int(c.GetspiderSize())
}

func SetRequestMiddleWares(c crawler, rmw RequestMiddleware) {
	c.RegisterRequestMiddleware(&rmw)
}

func SetResponseMiddleware(c crawler, rwm ResponseMiddleware) {
	c.RegisterResponseMiddleware(&rwm)
}

func SetStartingFunction(c crawler, f *func() []*BaseRequestObj) {
	c.RegisterStartingRequest(f)
}

func Crawl(c crawler) {
	c.StartCrawl()
}

func ShutDown(c crawler) {
	c.ShutDown()
}
