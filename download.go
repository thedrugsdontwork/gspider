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
				for _, rmw := range c.requestMiddlewares {
					if req == nil {
						break
					}
					req = (*rmw).process(req)
				}
				if req == nil {
					atomic.AddInt32(&c.reqIngCounter, -1)
					continue
				}
				c.logger.Debug("Start " + req.Method + " url:" + req.URL)
				res, err := req.Request.Execute(req.Method, req.URL)
				if err != nil {
					c.logger.Error(err.Error())
					panic(err)
				}
				var response *BaseResponseObj = AssembleResponse(res, req)
				for _, rmw := range c.responseMiddlewares {

					response = (*rmw).process(response)
					if response == nil {
						break
					}
				}
				if response != nil {
					c.submitResponse(response)

				}
				atomic.AddInt32(&c.reqIngCounter, -1)
			}
			c.wg.Done()
		}()
	}
}
