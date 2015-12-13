package pchome

import (
	"net/http"
	"log"

	"github.com/a2n/alu"
)

// PChome 服務結構。
type Service struct {
	Key string
	Logger *log.Logger
}

// PChome 存取點網址。
const (
	ENDPOINT = "http://myname.pchome.com.tw/manage"
)

// 記錄。
var logger *log.Logger

// 取得服務。
func NewService(key string) *Service {
	if len(key) == 0 {
		logger.Fatalf("%s has empty key.", alu.Caller())
	}

	return &Service {
		Key: key,
		Logger: alu.NewLogger("log"),
	}
}

// 取得 zone 服務。
func (s *Service) NewZoneService() *ZoneService {
	return &ZoneService {
		Service: s,
	}
}

// 取得 NS 服務。
func (s *Service) NewNSService() *NSService {
	return &NSService {
		Service: s,
	}
}

// 取得 DNSSEC 服務。
func (s *Service) NewDNSSECService() *DNSSECService {
	return &DNSSECService {
		Service: s,
	}
}

// 設定 cookie 內容。
func (s *Service)SetCookie(req *http.Request) {
	if req == nil {
		return
	}

	c := &http.Cookie {
		Name: "loginkuser",
		Value: s.Key,
	}

	req.AddCookie(c)
}
