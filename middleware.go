package GSpider

type Middleware interface {
	GetName() string
}
type RequestMiddleware interface {
	Middleware
	Process(crawler CrawlerInterface, req *BaseRequestObj) *MiddlewareResult
}
type ResponseMiddleware interface {
	Middleware
	Process(crawler CrawlerInterface, res *BaseResponseObj) *MiddlewareResult
}
type MiddlewareResult struct {
	Res *BaseResponseObj
	Req *BaseRequestObj
}

type BaseReuqestMiddleware struct {
}
