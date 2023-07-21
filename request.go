package GSpider

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"net/http"
	"strings"
)

type RequestInfo struct {
	Url      string
	Headers  map[string]string
	Cookies  map[string]string
	Method   string
	Proxy    map[string]string
	Callback func(obj *BaseResponseObj) []*BaseRequestObj
	Errback  func(obj *BaseResponseObj) []*BaseRequestObj
}
type RequestHandle interface {
	AssembleRequest() *BaseRequestObj
}
type BaseRequestObj struct {
	resty.Request
	Proxy      map[string]string
	Callback   func(obj *BaseResponseObj) []*BaseRequestObj
	Errback    func(obj *BaseResponseObj) []*BaseRequestObj
	RetryTimes int
}

var Methods = []string{http.MethodGet, http.MethodDelete, http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodPost, http.MethodPut, http.MethodTrace, http.MethodPatch}

/**
set proxy
*/
func (obj *RequestInfo) AssembleRequest() *BaseRequestObj {

	var client *resty.Client = resty.New()
	if obj.Proxy != nil {
		client.SetProxy(obj.Proxy["http"])
	}

	var req *resty.Request = client.R()
	var request BaseRequestObj = BaseRequestObj{
		Request: *req,
	}
	if obj.Method == "" {
		request.Method = http.MethodGet
	} else {
		_, found := Find(Methods, strings.ToUpper(obj.Method))
		if !found {
			request.Method = http.MethodGet
		} else {
			request.Method = strings.ToUpper(obj.Method)

		}
	}
	request.URL = obj.Url
	request.Method = obj.Method
	request.Proxy = obj.Proxy
	request.Callback = obj.Callback
	request.Errback = obj.Errback
	request.Header = http.Header{}
	request.RetryTimes = 0
	for key, value := range obj.Headers {
		request.Header.Set(key, value)
	}
	for key, value := range obj.Cookies {
		cookie := http.Cookie{}
		cookie.Name = key
		cookie.Value = value
		cookie.Path = "/"
		request.Cookies = append(request.Cookies, &cookie)
	}
	return &request
}
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}
func TestAssembleRequest() *BaseRequestObj {
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
		return nil
	}
	errback := func(obj *BaseResponseObj) []*BaseRequestObj {
		fmt.Printf("Response fail code:%d \n", obj.StatusCode)
		fmt.Printf("Url is :%s\n", obj.Url)
		fmt.Printf("Message is :%s\n", obj.Content)
		for key, val := range tmp {
			fmt.Printf("%s:%s\n", key, val)
		}
		fmt.Println("end\n")
		return nil
	}
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
	requestObj.Errback = errback
	requestObj.Headers = headers
	requestObj.Cookies = make(map[string]string, 0)
	requestObj.Proxy = map[string]string{"http": "http://127.0.0.1:8888"}
	//requestObj.Method = "GET"
	requestObj.Url = "https://sci.bban.top/pdf/10.3390/md16040118.pdf?download=true"
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
	return request
}
