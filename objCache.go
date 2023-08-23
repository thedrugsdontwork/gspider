package GSpider

/*
中间缓冲模块，防止请求或者响应的传递导致阻塞
*/

type reqCache interface {
	Load() *BaseRequestObj
	Store(obj *BaseRequestObj)
	CurSize() int
}

type resCache interface {
	Load() *BaseResponseObj
	Store(obj *BaseResponseObj)
	CurSize() int
}
