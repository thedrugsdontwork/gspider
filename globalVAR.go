package GSpider

//crawler

type _signal int

const (
	STOP    _signal = 0
	RUNNING _signal = 1
	WAITING _signal = 2
)

//request
const (
	DEFAULT_RETRY_TIMES = 5
)

//base request cache
const (
	DEFAULT_MAX_SIZE             = 1000
	DEFAULT_PAGE_SIZE            = 100
	DEFAULT_REQ_CACHE_LOCAL_FILE = "requestlocalcache"
)

//header
const (
	A_IM                           = "A-IM"
	ACCEPT                         = "Accept"
	ACCEPT_CHARSET                 = "Accept-Charset"
	ACCEPT_ENCODING                = "Accept-Encoding"
	ACCEPT_LANGUAGE                = "Accept-Language"
	ACCEPT_DATETIME                = "Accept-Datetime"
	ACCESS_CONTROL_REQUEST_METHOD  = "Access-Control-Request-Method"
	ACCESS_CONTROL_REQUEST_HEADERS = "Access-Control-Request-Headers"
	AUTHORIZATION                  = "Authorization"
	CACHE_CONTROL                  = "Cache-Control"
	CONNECTION                     = "Connection"
	CONTENT_LENGTH                 = "Content-Length"
	CONTENT_TYPE                   = "Content-Type"
	COOKIE                         = "Cookie"
	DATE                           = "Date"
	EXPECT                         = "Expect"
	FORWARDED                      = "Forwarded"
	FROM                           = "From"
	HOST                           = "Host"
	IF_MATCH                       = "If-Match"
	IF_MODIFIED_SINCE              = "If-Modified-Since"
	IF_NONE_MATCH                  = "If-None-Match"
	IF_RANGE                       = "If-Range"
	IF_UNMODIFIED_SINCE            = "If-Unmodified-Since"
	MAX_FORWARDS                   = "Max-Forwards"
	ORIGIN                         = "Origin"
	PRAGMA                         = "Pragma"
	PROXY_AUTHORIZATION            = "Proxy-Authorization"
	RANGE                          = "Range"
	REFERER                        = "Referer"
	TE                             = "TE"
	USER_AGENT                     = "User-Agent"
	UPGRADE                        = "Upgrade"
	VIA                            = "Via"
	WARNING                        = "Warning"
	DNT                            = "Dnt"
	X_REQUESTED_WITH               = "X-Requested-With"
	X_CSRF_TOKEN                   = "X-CSRF-Token"
)
