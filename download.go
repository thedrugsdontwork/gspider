package GSpider

import (
	"fmt"
	"sync/atomic"
)

/*
	crawl core
*/
func (c *Crawler) download() {
	c.wg.Add(int(c.spiderSize))
	for i := 0; i < int(c.spiderSize); i++ {
		go func() {
			fmt.Println("Starting wait for get reuqest")
			for req := range c.requests {
				atomic.AddInt32(&c.requestsCurSize, -1)
				atomic.AddInt32(&c.reqIngCounter, 1)
				atomic.AddInt32(&c.rCounter, 1)

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
				res, err := req.Request.Execute(req.Method, req.URL)
				if err != nil {
					fmt.Println(err)
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
					atomic.AddInt32(&c.responseCurSize, 1)
					//c.responses <- response
					c.resCache.Store(response)
				}
				atomic.AddInt32(&c.reqIngCounter, -1)
			}
			c.wg.Done()
		}()
	}
}
