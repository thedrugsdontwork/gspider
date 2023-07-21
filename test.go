package GSpider

import (
	"fmt"
	"io/ioutil"
)

//type Crawler struct {
//	poolSize            int
//	crawlerName         string
//	requestMiddlewares  []*RequestMiddleware
//	responseMiddlewares []*ResponseMiddleware
//	reqMSize            int
//	resMSize            int
//	startFunc           *func() []*BaseRequestObj
//	requests            chan *BaseRequestObj
//	requestsCurSize     int32
//	responses           chan *BaseResponseObj
//	responseCurSize     int32
//	signals             []_signal
//	rCounter            int32
//	reqIngCounter       int32
//	parIngCounter       int32
//	wg                  sync.WaitGroup
//}
func StartRequest() []*BaseRequestObj {
	requests := []*BaseRequestObj{}
	for i := 0; i < 100; i++ {
		var tmp map[string]string = map[string]string{
			"asda": "asdad",
			"name": "monkey",
		}
		tmp["asdasdad001"] = "sadasdadadaaaa"
		callback := func(obj *BaseResponseObj) []*BaseRequestObj {
			fmt.Printf("Response successfully code:%d \n", obj.StatusCode)
			fmt.Printf("Url is :%s\n", obj.Url)

			err := ioutil.WriteFile("test.pdf", obj.Content, 755)
			if err != nil {
				panic(err)
			}
			//fmt.Printf("Response content is :%s", obj.Content)
			fmt.Printf("Get local var :\n")
			for key, val := range tmp {
				fmt.Printf("%s:%s\n", key, val)
			}
			fmt.Println("end\n")
			fmt.Printf("Add key local and val yes!\n")
			tmp["local"] = "yes"
			fmt.Printf("%s:%s\n", "local", tmp["local"])
			if obj.Request.RetryTimes < 3 {
				fmt.Printf("Retry times is %d start retry", obj.Request.RetryTimes)
				obj.Request.RetryTimes += 1
				requests := []*BaseRequestObj{obj.Request}

				return requests
			}
			fmt.Println("\nStop sumit request!!!")
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
func TestForCrawler() {
	var crawler = Crawler{}
	crawler.Init("TestForCrawler", 4)
	var f = StartRequest
	crawler.requests = make(chan *BaseRequestObj)
	crawler.responses = make(chan *BaseResponseObj)
	crawler.RegisterStartingRequest(&f)
	fmt.Printf("Start crawl\n")
	crawler.StartCrawl()

}
