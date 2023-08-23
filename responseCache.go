package GSpider

/*
	采集过程中响应缓存,分为两级:
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

type BaseResCacheObj struct {
	/*fifo*/

	next *BaseResCacheObj
	val  *BaseResponseObj
}

type BaseResCache struct {
	size             int
	head             *BaseResCacheObj
	m                sync.Mutex
	name             string
	pageSize         int
	totalSize        int
	cacheSize        int
	maxCacheSize     int
	localStorageFile string //default prj root dir
	fileHandle       *os.File
	loadOffset       int64
	storageOffset    int64
	loadBuf          []byte
	logger           *zap.Logger
}

func (brc *BaseResCache) Load() *BaseResponseObj {
	/*
		return the last object and reduce the size
	*/

	var res *BaseResponseObj
	brc.m.Lock()
	if brc.cacheSize == 0 && brc.totalSize > 0 {
		brc.loadFromLocal()
	}
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
	if brc.cacheSize >= brc.maxCacheSize {
		brc.storeLocal()
	}
	var tmp = BaseResCacheObj{
		next: brc.head,
		val:  obj,
	}
	brc.head = &tmp
	brc.size++
	brc.m.Unlock()
}

func (brc *BaseResCache) CurSize() int {
	return brc.totalSize
}

func (brc *BaseResCache) storeLocal() {
	//将对象写入本地文件
	brc.logger.Debug("Write response to local file:" + brc.localStorageFile)
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

func (brc *BaseResCache) loadFromLocal() {
	//从本地文件中加载
	brc.logger.Debug("Load response from local file:" + brc.localStorageFile)
	m.Lock()

	brc.fileHandle.Seek(brc.loadOffset, 0)

	var dec = gob.NewDecoder(brc.fileHandle)
	// buffer default size=4096
	for i := 0; i < brc.pageSize; i++ {
		resObj := BaseResponseObj{}
		dec.Decode(&resObj)
		tmp := BaseResCacheObj{
			next: brc.head,
			val:  &resObj,
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

func (brc *BaseResCache) close(crawler CrawlerInterface) {
	brc.logger.Info("Start close response cache.")
	os.Remove(brc.localStorageFile)
	brc.logger.Info("Response cache close completely", zap.String("name", brc.name))
}

func NewBaseResCache(crawler CrawlerInterface, pageSize int, maxSize int, storageFile string) *BaseResCache {
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
	var rCache = BaseResCache{
		name:             "baseResponseCache",
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
