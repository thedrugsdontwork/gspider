package GSpider

import (
	"github.com/go-resty/resty/v2"
	"net/http"
	"sync"
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

type BaseResCacheObj struct {
	/*fifo*/

	next *BaseResCacheObj
	val  *BaseResponseObj
}
type BaseResCache struct {
	size int
	head *BaseResCacheObj
	m    sync.Mutex
}

func (brc *BaseResCache) Load() *BaseResponseObj {
	/*
		return the last object and reduce the size
	*/

	var res *BaseResponseObj
	brc.m.Lock()
	if brc.size > 0 {
		var tmp *BaseResCacheObj = brc.head
		var pre *BaseResCacheObj = brc.head
		for i := 1; i < brc.size; i++ {
			pre = tmp
			tmp = tmp.next
		}
		if brc.size == 1 {
			brc.head = nil
		} else {
			pre.next = nil
		}
		brc.size--
		res = tmp.val
	}
	brc.m.Unlock()
	return res
}
func (brc *BaseResCache) Store(obj *BaseResponseObj) {
	/*
		store the obj in head and add to the size
	*/
	brc.m.Lock()
	var tmp = BaseResCacheObj{
		next: brc.head,
		val:  obj,
	}
	brc.head = &tmp
	brc.size++
	brc.m.Unlock()
}
func (brc *BaseResCache) CurSize() int {
	return brc.size
}
