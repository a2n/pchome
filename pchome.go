package pchome

import (
	"net/http"
	"log"

	"github.com/a2n/alu"
)

type Service struct {
	Key string
	Logger *log.Logger
}

const (
	ENDPOINT = "http://myname.pchome.com.tw/manage"
)

var logger *log.Logger
func NewService(key string) *Service {
	if len(key) == 0 {
		logger.Fatalf("%s has empty key.", alu.Caller())
	}

	return &Service {
		Key: key,
		Logger: alu.NewLogger("log"),
	}
}

func (s *Service) NewZoneService() *ZoneService {
	return &ZoneService {
		Service: s,
	}
}

func (s *Service) NewNSService() *NSService {
	return &NSService {
		Service: s,
	}
}

func (s *Service) NewDNSSECService() *DNSSECService {
	return &DNSSECService {
		Service: s,
	}
}

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
