package GSpider

import (
	"github.com/go-resty/resty/v2"
	"net/http"
)

type ResponseHandle interface {
	AssembleResponse(obj interface{}) *BaseResponseObj
}
type BaseResponseObj struct {
	http.Response
	Url        string
	StatusCode int
	Content    []byte
	Size       int64
	Request    *BaseRequestObj
}

func AssembleResponse(res *resty.Response, req *BaseRequestObj) *BaseResponseObj {
	var response = BaseResponseObj{}
	response.Request = req
	response.Url = res.Request.URL
	response.StatusCode = res.StatusCode()
	response.Size = res.Size()
	response.Content = res.Body()

	return &response
}
