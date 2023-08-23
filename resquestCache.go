package GSpider

/*
	采集过程中请求缓存，分为两级:
		1.内存cache
		2.本地文件cache
*/

import (
	"encoding/gob"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type BaseReqCacheObj struct {
	/*fifo*/

	next *BaseReqCacheObj
	val  *BaseRequestObj
}

type BaseReqCache struct {
	name             string
	pageSize         int
	totalSize        int
	cacheSize        int
	head             *BaseReqCacheObj
	m                sync.Mutex
	maxCacheSize     int
	localStorageFile string //default prj root dir
	fileHandle       *os.File
	loadOffset       int64
	storageOffset    int64
	loadBuf          []byte
	logger           *zap.Logger
}

func (brc *BaseReqCache) Load() *BaseRequestObj {
	/*
		return the last object and reduce the size
	*/
	var res *BaseRequestObj = nil
	brc.m.Lock()
	if brc.cacheSize == 0 && brc.totalSize > 0 {
		brc.loadFromLocal()
	}
	if brc.cacheSize > 0 {
		var tmp *BaseReqCacheObj = brc.head
		var pre *BaseReqCacheObj = brc.head

		for tmp.next != nil {
			pre = tmp
			tmp = tmp.next
		}
		if brc.cacheSize == 1 {
			brc.head = nil
		} else {
			pre.next = nil
		}
		brc.cacheSize--
		brc.totalSize--
		res = tmp.val
	}
	brc.m.Unlock()
	return res
}

func (brc *BaseReqCache) Store(obj *BaseRequestObj) {
	/*
		store the obj in head and add to the size
	*/
	brc.m.Lock()
	if brc.cacheSize >= brc.maxCacheSize {
		brc.storeLocal()
	}
	var tmp = BaseReqCacheObj{
		next: brc.head,
		val:  obj,
	}
	brc.head = &tmp
	brc.cacheSize++
	brc.totalSize++
	brc.m.Unlock()
}

func (brc *BaseReqCache) CurSize() int {
	return brc.totalSize
}

func (brc *BaseReqCache) storeLocal() {
	//将对象写入本地文件
	brc.logger.Debug("Write request to local file:" + brc.localStorageFile)
	m.Lock()
	brc.fileHandle.Seek(brc.storageOffset, 0)
	var enc = gob.NewEncoder(brc.fileHandle)
	for i := 0; i < brc.pageSize; i++ {
		enc.Encode(brc.head.val)
		brc.head = brc.head.next
		brc.cacheSize--
	}
	offset, err := brc.fileHandle.Seek(0, io.SeekCurrent)
	if err != nil {
		panic("Get file offset error:" + err.Error())
	}
	//补齐
	if offset%4096 > 0 {
		padding := 4096 - (offset % 4096)
		bs := make([]byte, padding)
		brc.fileHandle.Write(bs)
		offset += padding
	}
	brc.storageOffset = offset
	m.Unlock()
}

func (brc *BaseReqCache) loadFromLocal() {
	//从本地文件中加载
	brc.logger.Debug("Load request from local file:" + brc.localStorageFile)
	m.Lock()

	brc.fileHandle.Seek(brc.loadOffset, 0)

	var dec = gob.NewDecoder(brc.fileHandle)
	// buffer default size=4096
	for i := 0; i < brc.pageSize; i++ {
		reqObj := BaseRequestObj{}
		dec.Decode(&reqObj)
		tmp := BaseReqCacheObj{
			next: brc.head,
			val:  &reqObj,
		}
		brc.head = &tmp
		brc.cacheSize++
		brc.fileHandle.Seek(0, io.SeekCurrent)
	}
	offset, err := brc.fileHandle.Seek(0, io.SeekCurrent)
	if err != nil {
		panic("Get file offset error:" + err.Error())
	}
	brc.loadOffset = offset
	m.Unlock()
}

func (brc *BaseReqCache) close(crawler CrawlerInterface) {
	brc.logger.Info("Start close request cache.")
	os.Remove(brc.localStorageFile)
	brc.logger.Info("Request cache close completely", zap.String("name", brc.name))
}

func NewBaseReqCache(crawler CrawlerInterface, pageSize int, maxSize int, storageFile string) *BaseReqCache {
	//get prj root dir
	var file string

	if storageFile == "" {
		dir, error := filepath.Abs("./")
		if error != nil {
			panic("Get abs path error:" + error.Error())
			return nil
		}
		file = filepath.Join(dir, DEFAULT_REQ_CACHE_LOCAL_FILE)
	} else {
		file, _ = filepath.Abs(storageFile)
	}
	if pageSize > maxSize {
		panic("PageSize should be less than MaxSize")
	}
	if pageSize <= 0 {
		pageSize = DEFAULT_PAGE_SIZE
	}
	if maxSize <= 0 {
		maxSize = DEFAULT_MAX_SIZE
	}
	var fp, _ = os.Create(file)
	var rCache = BaseReqCache{
		name:             "baseRequestCache",
		pageSize:         pageSize,
		localStorageFile: file,
		m:                sync.Mutex{},
		fileHandle:       fp,
		maxCacheSize:     maxSize,
		loadOffset:       0,
		storageOffset:    0,
		logger:           crawler.GetLogger(),
	}
	f := rCache.close
	//订阅关闭事件
	crawler.Subscribe(STOP, &f)
	return &rCache
}
