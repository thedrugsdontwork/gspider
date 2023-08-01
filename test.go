package GSpider

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"sync"
)

var fp, _ = os.Create("test.json")
var m = sync.Mutex{}

func StartRequest() []*BaseRequestObj {
	requests := []*BaseRequestObj{}
	logger, _ := zap.NewProduction()

	for i := 0; i < 100; i++ {
		var tmp map[string]string = map[string]string{
			"asda": "asdad",
			"name": "monkey",
		}
		tmp["asdasdad001"] = "sadasdadadaaaa"
		callback := func(obj *BaseResponseObj) []*BaseRequestObj {
			defer fp.Sync()
			defer logger.Sync()
			logger.Info("Response successfully code:%d \n", zap.Int32("statusCode", int32(obj.StatusCode)))
			logger.Info("Url is : " + obj.Url)
			m.Lock()
			fp.Write(obj.Content)
			logger.Info("Response content ", zap.String("content", "write file ./test.json"))
			m.Unlock()
			logger.Info("Response content ", zap.ByteString("content", obj.Content))

			if obj.Request.RetryTimes < 3 {
				logger.Warn("Retry times is %d start retry", zap.Int32("retryTimes", int32(obj.Request.RetryTimes)))
				obj.Request.RetryTimes += 1
				requests := []*BaseRequestObj{obj.Request}

				return requests
			}
			logger.Info("Stop sumit request!!!")
			return nil
		}
		//errback := func(obj *BaseResponseObj) {
		//	fmt.Printf("Response fail code:%d \n", obj.StatusCode)
		//	fmt.Printf("Url is :%s\n", obj.Url)
		//	fmt.Printf("Message is :%s\n", obj.Content)
		//	for key, val := range tmp {
		//		fmt.Printf("%s:%s\n", key, val)
		//	}
		//	fmt.Println("end\n")
		//}
		var headers map[string]string = make(map[string]string)

		headers["accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9"
		headers["accept-encoding"] = "gzip, deflate"
		headers["accept-language"] = "zh,zh-CN;q=0.9"
		headers["cache-control"] = "no-cache"
		headers["origin"] = "https://www.baidu.com"
		headers["pragma"] = "no-cache"
		headers["referer"] = "https://www.baidu.com"
		headers["user-agent"] = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/38.0.2125.122 Safari/537.36 SE 2.X MetaSr 1.0"
		var requestObj RequestInfo = RequestInfo{}
		requestObj.Callback = callback
		//requestObj.Errback = errback
		requestObj.Headers = headers
		requestObj.Cookies = make(map[string]string, 0)
		//requestObj.Proxy = map[string]string{"http": "http://127.0.0.1:8888"}
		requestObj.Method = "GET"
		requestObj.Url = "http://127.0.0.1/get"
		//https://sci.bban.top/pdf/10.3390/md16040118.pdf?download=true
		var request = requestObj.AssembleRequest()
		fmt.Println("header is :\n")
		for key, value := range request.Header {
			fmt.Printf("%s:%s\n", key, value)
		}
		fmt.Println("\ncookies is :\n")
		for idx := range request.Cookies {
			cookie := request.Cookies[idx]
			fmt.Printf("%s:%s\n", cookie.Name, cookie.Value)
			//fmt.Println("%s:%s", cookie.Name, cookie.Value)
		}
		fmt.Println("proxies is :", request.Proxy)

		requests = append(requests, request)
	}
	return requests
}
func process() *BaseResponseObj {
	var res *BaseResponseObj
	return res
}

type msdasdMW struct {
	counter int32
}

func (obj *msdasdMW) process(r *BaseResponseObj) *BaseResponseObj {
	var res *BaseResponseObj
	res = r
	obj.counter++
	fmt.Printf("Current process in the res middle ware process total %d", obj.counter)
	return res
}
func TestForCrawler() {

	var crawler = Crawler{}
	var mw *msdasdMW = &msdasdMW{
		counter: 0,
	}
	//crawler.Init("TestForCrawler", 4, 4)
	SpiderInit(&crawler, "TestForCrawler", 4, 4)
	SetResponseMiddleware(&crawler, mw)
	//SetRequestMiddleWares(&crawler,)
	var f = StartRequest
	SetStartingFunction(&crawler, &f)
	Crawl(&crawler)
	for crawler.signal != STOP {

	}

	fp.Close()

}
