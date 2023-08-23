package GSpider

/*
代理中间件，用于对请求添加代理
*/
type ProxyMiddleware struct {
	name string
}

func (pmw *ProxyMiddleware) Process(crawler CrawlerInterface, req *BaseRequestObj) *BaseRequestObj {

	return req

}
