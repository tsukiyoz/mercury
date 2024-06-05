package logger

type AccessLog struct {
	Method   string
	Url      string
	Duration string
	ReqBody  string
	RespBody string
	Status   int
}

func limitString(val string) string {
	if len(val) > 1024 {
		val = val[:1024] + "..."
	}
	return val
}

func limitByte(val []byte) []byte {
	if len(val) > 1024 {
		val = append(val, []byte("...")...)
	}
	return val
}

func NewAccessLog(method string, url string) *AccessLog {
	return &AccessLog{
		Method: method,
		Url:    url,
	}
}
