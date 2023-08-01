package GSpider

import (
	"go.uber.org/zap"
	"sync/atomic"
)

func (c *Crawler) parse() {
	c.wg.Add(int(c.parserSize))
	for i := 0; i < int(c.parserSize); i++ {

		go func() {
			defer c.logger.Sync()
			var reqs []*BaseRequestObj
			for res := range c.responses {
				atomic.AddInt32(&c.responseCurSize, -1)
				atomic.AddInt32(&c.parIngCounter, 1)
				if res.StatusCode < 400 {
					c.logger.Debug("Url:"+res.Url+" response success", zap.Int32("statusCode", int32(res.StatusCode)))
					if res.Request.Callback != nil {
						reqs = res.Request.Callback(res)
					}
				} else {
					c.logger.Debug("Url:"+res.Url+" response error", zap.Int32("statusCode", int32(res.StatusCode)))
					if res.Request.Errback != nil {
						reqs = res.Request.Errback(res)
					}
				}
				if reqs != nil {
					c.submitRequest(reqs)
				}
				atomic.AddInt32(&c.parIngCounter, -1)
			}
			c.wg.Done()
		}()
	}
}
