package GSpider

import (
	"sync/atomic"
)

/*
	crawl core
*/
func (c *Crawler) download() {
	c.wg.Add(int(c.spiderSize))
	for i := 0; i < int(c.spiderSize); i++ {
		go func() {
			defer c.logger.Sync()
			for req := range c.requests {
				atomic.AddInt32(&c.requestsCurSize, -1)
				atomic.AddInt32(&c.reqIngCounter, 1)
				/*请求中间件处理*/
				for _, rmw := range c.requestMiddlewares {
					if req == nil {
						break
					}
					mRes := (*rmw).Process(c, req)
					if mRes.Req == nil {
						c.logger.Warn("Request [" + req.Method + "] Url:" + req.URL + "process get nil from request middle:" + (*rmw).GetName())
						break
					}
					req = mRes.Req
				}
				if req == nil {
					atomic.AddInt32(&c.reqIngCounter, -1)
					continue
				}
				/*发起请求*/
				c.logger.Debug("Start [" + req.Method + "] url:" + req.URL)
				res, err := req.Request.Execute(req.Method, req.URL)
				if err != nil {
					c.logger.Error(err.Error())
					panic(err)
				}
				var response *BaseResponseObj = AssembleResponse(res, req)

				/*响应中间件处理：返回
				请求 若返回结构中含有请求，则将请求放入请求队列，
				响应 若返回结构中含有响应，则继续中间件处理
				nil 若返回结构中请求响应都为空时，停止中间件处理
				*/
				for _, rmw := range c.responseMiddlewares {

					mRes := (*rmw).Process(c, response)

					if mRes.Req != nil {
						c.submitRequest([]*BaseRequestObj{mRes.Req})
					}
					response = mRes.Res
					if response == nil {
						break
					}
				}
				/*响应中间件处理完成后将响应发送至响应队列*/
				if response != nil {
					c.submitResponse(response)
				}
				atomic.AddInt32(&c.reqIngCounter, -1)
			}
			c.wg.Done()
		}()
	}
}
