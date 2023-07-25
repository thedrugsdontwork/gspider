package GSpider

/*
中间缓冲模块，防止请求或者响应的传递导致阻塞
*/

type cache interface {
	Load() interface{}
	Store(obj interface{})
	CurSize() int
}
