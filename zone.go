package pchome

import (
	"net/http"
	"io/ioutil"
	"regexp"

	"github.com/a2n/alu"
)

type Zone struct {
	NS NS
	DNSSEC []DNSSEC
}

type ZoneService struct {
	Service *Service
	cs *ConfigService
	config Config
	zone string
}

func (zs *ZoneService) List() *ZoneListCall {
	return &ZoneListCall {
		Service: zs.Service,
	}
}

type ZoneListCall struct {
	Service *Service
}

func (zlc *ZoneListCall) Do() map[string]Zone {
	urlstr := ENDPOINT + "/index.htm"
	req, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		logger.Fatalf("%s creates request failed, %s.", alu.Caller(), err.Error())
	}
	zlc.Service.SetCookie(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("%s requesting failed, %s.", alu.Caller(), err.Error())
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("%s requesting failed, %s.", alu.Caller(), err.Error())
	}
	resp.Body.Close()
	return zlc.Parse(b)
}

func (zlc *ZoneListCall) Parse(raw []byte) map[string]Zone {
	zones := make(map[string]Zone)
	if len(raw) == 0 {
		logger.Printf("%s has empty raw.", alu.Caller())
		return zones
	}

	re := regexp.MustCompile(`\?dn=(.*)">進入`)
	if !re.Match(raw) {
	    return zones
	}
	ms := re.FindAllStringSubmatch(string(raw), -1)
	for _, v := range ms {
		zones[v[1]] = Zone{}
	}

	return zones
}
