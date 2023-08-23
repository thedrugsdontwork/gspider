package GSpider

//todo 1.添加日志收集器 done
//todo 2.优化cache阈值缓存本地 done 50% 响应cache暂未本地化
//todo 3.添加下载统计（定时）done
//todo 4.添加默认中间件（代理，随机ua）done
//todo 5.与数据库一起综合测试
//todo 6.添加全局发布订阅，用于爬虫信号变更时各个组件的响应 done
//todo 7.状态获取以及状态更改加锁 done
//todo 8.由于部分obj无法序列化，请求以及响应的本地缓冲无法使用，需要将请求响应object重构，使其可接受序列化

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

type CrawlerInterface interface {
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
	RegisterRequestMiddleware(middleWare RequestMiddleware)
	/*
		注册响应前置处理中间件
	*/
	RegisterResponseMiddleware(middleWare ResponseMiddleware)
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

	SetConfFromMap(confs map[string]interface{})
	/*添加配置*/
	SetConf(key string, value interface{})
	/*获取配置*/
	GetConf(key string) interface{}

	GetConfAsStr(key string) string

	GetConfAsInt(key string) int

	GetLogger() *zap.Logger
	/*订阅事件现只支持状态变化*/
	Subscribe(status _signal, f *func(crawler CrawlerInterface))
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
	//通道当前大小
	requestsCurSize int32
	responses       chan *BaseResponseObj
	responseCurSize int32
	signals         []_signal
	rCounter        int32
	//当前处理线程
	reqIngCounter int32
	parIngCounter int32
	wg            sync.WaitGroup
	signal        _signal
	reqCache      reqCache
	resCache      resCache
	logger        *zap.Logger
	startTime     time.Time
	config        map[string]interface{}
	Subscribes    map[_signal][]*func(crawler CrawlerInterface)
	statLock      sync.Mutex
}

func (c *Crawler) Init(crawlerName string, spiderSize int32, parserSize int32) {
	c.spiderSize = spiderSize
	c.parserSize = parserSize
	c.crawlerName = crawlerName
	c.requests = make(chan *BaseRequestObj, spiderSize)
	c.responses = make(chan *BaseResponseObj, parserSize)
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

	//get cache use default config
	c.reqCache = NewBaseReqCache(c, 2, 3, "")
	c.resCache = NewBaseResCache(c, 2, 3, "")

	//register default middleware
	c.RegisterRequestMiddleware(NewRandomUAMiddleware(c, []string{}))
	c.RegisterResponseMiddleware(NewRetryMiddleware(c, 3))
}

func (c *Crawler) GetspiderSize() int32 {
	return c.spiderSize
}

func (c *Crawler) RegisterRequestMiddleware(middleWare RequestMiddleware) {
	c.logger.Info("Start register request middleware:" + middleWare.GetName())
	c.requestMiddlewares = append(c.requestMiddlewares, &middleWare)
	c.reqMSize++
	c.logger.Info("Complete register request middleware:" + middleWare.GetName())
}

func (c *Crawler) RegisterResponseMiddleware(middleWare ResponseMiddleware) {
	c.logger.Info("Start register response middleware:" + middleWare.GetName())
	c.responseMiddlewares = append(c.responseMiddlewares, &middleWare)
	c.resMSize++
	c.logger.Info("Complete register response middleware:" + middleWare.GetName())
}

func (c *Crawler) RegisterStartingRequest(f *(func() []*BaseRequestObj)) {
	if c.startFunc != nil {
		panic("Start function already registered.")
	}
	c.startFunc = f
}

//所有请求提交
func (c *Crawler) submitRequest(reqs []*BaseRequestObj) {

	for _, req := range reqs {
		if req != nil {
			c.reqCache.Store(req)
			atomic.AddInt32(&c.rCounter, 1)
			atomic.AddInt32(&c.requestsCurSize, 1)
		}
	}
}

//所有响应提交
func (c *Crawler) submitResponse(res *BaseResponseObj) {

	if res != nil {
		c.resCache.Store(res)
		atomic.AddInt32(&c.responseCurSize, 1)
	}
}

func (c *Crawler) SetConfFromMap(conf map[string]interface{}) {
	for key, val := range conf {
		c.config[key] = val
	}
}

func (c *Crawler) SetConf(key string, value interface{}) {
	c.config[key] = value
}

func (c *Crawler) GetConf(key string) interface{} {
	return c.config[key]
}

func (c *Crawler) GetConfAsStr(key string) string {
	var val = c.config[key]
	if str, ok := val.(string); ok {
		return str
	} else {
		panic("This key:'" + key + "' value is not string ,[" + reflect.TypeOf(val).String() + "]")
	}
}

func (c *Crawler) GetConfAsInt(key string) int {
	var val = c.config[key]
	if v, ok := val.(int); ok {
		return v
	} else {
		panic("This key:'" + key + "' value is not int ,[" + reflect.TypeOf(val).String() + "]")
	}
}

func (c *Crawler) ChangeStatus(status _signal) {
	c.statLock.Lock()
	c.signal = status
	c.statLock.Unlock()
	//发布事件
	lis := c.Subscribes[status]
	if lis != nil {
		for _, f := range lis {
			(*f)(c)
		}
	}
}

func (c *Crawler) IsStop() bool {
	flag := false
	c.statLock.Lock()
	flag = c.signal == STOP
	c.statLock.Unlock()
	return flag
}

//开始采集
func (c *Crawler) StartCrawl() {
	defer c.logger.Sync()
	c.logger.Info("Crawler has been running.", zap.String("name", c.crawlerName))
	c.wg.Add(4)
	c.startTime = time.Now()
	c.ChangeStatus(WAITING)
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
	c.ChangeStatus(RUNNING)
	c.logger.Info("Start function total submit reuqest ", zap.Int32("start request size", atomic.LoadInt32(&c.rCounter)))
	/*use to submit req to channel*/
	go func() {
		for !c.IsStop() {
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
		for !c.IsStop() {
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
		for c.reqCache.CurSize() != 0 || c.resCache.CurSize() != 0 || atomic.LoadInt32(&c.reqIngCounter) != 0 || atomic.LoadInt32(&c.parIngCounter) != 0 || atomic.LoadInt32(&c.requestsCurSize) != 0 || atomic.LoadInt32(&c.responseCurSize) != 0 {
			time.Sleep(1 * time.Second)
		}

		c.ShutDown()
		c.wg.Done()
	}()
	/*statistic all target*/
	go func() {
		defer c.logger.Sync()
		for !c.IsStop() {
			time.Sleep(1 * time.Second)
			now := time.Now()
			now.Sub(c.startTime)
			c.logger.Info("Crawler statistic", zap.String("name", c.crawlerName), zap.Int32("totalPage", atomic.LoadInt32(&c.rCounter)), zap.Float64("spend/s", now.Sub(c.startTime).Seconds()), zap.Int32("per3Sec/pages", atomic.LoadInt32(&c.rCounter)/int32(now.Sub(c.startTime).Seconds())))
		}
		c.wg.Done()
	}()
	c.wg.Wait()
	fmt.Printf("Program shutdown.")
}

func (c *Crawler) GetLogger() *zap.Logger {
	return c.logger
}

//停止
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
	c.ChangeStatus(STOP)
	c.logger.Info("Crawler "+c.crawlerName+" has been closed", zap.String("crawlerName", c.crawlerName))
}

func (c *Crawler) Subscribe(status _signal, f *func(crawler CrawlerInterface)) {
	if c.Subscribes == nil {
		c.Subscribes = map[_signal][]*func(crawler CrawlerInterface){}
	}
	lis := c.Subscribes[status]
	if lis == nil {
		lis = []*func(crawler CrawlerInterface){}
	}
	c.Subscribes[status] = append(lis, f)
}

/*
	用于装配爬虫
*/
func SpiderInit(c CrawlerInterface, crawlerName string, spiderSize int32, parserSize int32) {
	c.Init(crawlerName, spiderSize, parserSize)
}

func GetSpiderSize(c CrawlerInterface) int {
	return int(c.GetspiderSize())
}

func SetRequestMiddleWares(c CrawlerInterface, rmw RequestMiddleware) {
	c.RegisterRequestMiddleware(rmw)
}

func SetResponseMiddleware(c CrawlerInterface, rwm ResponseMiddleware) {
	c.RegisterResponseMiddleware(rwm)
}

func SetStartingFunction(c CrawlerInterface, f *func() []*BaseRequestObj) {
	c.RegisterStartingRequest(f)
}

func Crawl(c CrawlerInterface) {
	c.StartCrawl()
}

func SetConf(c CrawlerInterface, key string, value interface{}) {
	c.SetConf(key, value)
}

func GetConf(c CrawlerInterface, key string) interface{} {
	return c.GetConf(key)
}

func ShutDown(c CrawlerInterface) {
	c.ShutDown()
}
