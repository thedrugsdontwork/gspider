package GSpider

import (
	"go.uber.org/zap"
	"sync"
)

type RetryMiddleware struct {
	name       string `default:"RetryMiddleware"`
	retryTimes int
	lock       sync.Mutex
}

func (obj *RetryMiddleware) Process(crawler CrawlerInterface, res *BaseResponseObj) *MiddlewareResult {

	if res.StatusCode >= 400 && res.StatusCode < 500 {
		return &MiddlewareResult{Res: res}
	}

	if res.Request.RetryTimes < obj.retryTimes {
		obj.lock.Lock()
		res.Request.RetryTimes++
		obj.lock.Unlock()
		return &MiddlewareResult{Req: res.Request}
	}
	return &MiddlewareResult{Res: res}

}

func (obj *RetryMiddleware) GetName() string {
	return obj.name
}

func (obj *RetryMiddleware) close(crawler CrawlerInterface) {
	crawler.GetLogger().Info("Middleware closed", zap.String("name", obj.name))
}
func NewRetryMiddleware(crawler CrawlerInterface, retry int) *RetryMiddleware {

	mw := RetryMiddleware{
		name:       "RetryMiddleware",
		retryTimes: retry,
		lock:       sync.Mutex{},
	}
	if retry <= 0 {
		mw.retryTimes = DEFAULT_RETRY_TIMES
	}
	f := mw.close
	crawler.Subscribe(STOP, &f)
	return &mw
}
