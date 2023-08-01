package GSpider

type RequestMiddleware interface {
	process(req *BaseRequestObj) *BaseRequestObj
}
type ResponseMiddleware interface {
	process(res *BaseResponseObj) *BaseResponseObj
}

type BaseReuqestMiddleware struct {
}
