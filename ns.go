package pchome

import (
	"net/http"
	"io/ioutil"
	"regexp"
	"strings"
	"net/url"
	"errors"
	"strconv"

	"github.com/a2n/alu"
)

// Name server 結構。
type NS map[string]string

// NS 服務結構。
type NSService struct {
	Service *Service
	cs *ConfigService
	config Config
	zone string
}

// 添加 NS 記錄。
func (ns *NSService) Add(zone, name, ip string) error {
	ns.cs = NewConfigService()
	config, err := ns.cs.Read()
	if err != nil {
		return err
	}
	ns.config = config

	// Zone
	if _, ok := ns.config.Zones[zone]; !ok {
		logger.Fatal("%s has no such zone name, %s.", alu.Caller(), zone)
		return errors.New("No such zone name.")
	}
	ns.zone = zone

	if len(ns.config.Zones[ns.zone].NS) == 5 {
		logger.Fatal("%s, zone(%s) is reaching the max NS record count 5.", alu.Caller(), zone)
		return errors.New("The zone is reaching the max NS record count 5, delete some records first.")
	}

	// Name
	if _, ok := ns.config.Zones[ns.zone].NS[name]; ok {
		logger.Fatal("%s has duplicated host name, %s.", alu.Caller(), name)
		return errors.New("Duplicated host name.")
	}

	ns.config.Zones[ns.zone].NS[name] = ip
	ns.save()

	return nil
}

// 移除 NS 記錄。
func (ns *NSService) Delete(zone, name, ip string) error {
	ns.cs = NewConfigService()
	config, err := ns.cs.Read()
	if err != nil {
		return err
	}
	ns.config = config

	// Zone
	if _, ok := ns.config.Zones[zone]; !ok {
		logger.Fatal("%s has no matched zone name, %s.", alu.Caller(), zone)
		return errors.New("No matched zone name.")
	}
	ns.zone = zone

	// Name
	if _, ok := ns.config.Zones[ns.zone].NS[name]; !ok {
		logger.Fatal("%s no matched host name, %s.", alu.Caller(), name)
		return errors.New("No matched host name.")
	}

	// IP
	if ns.config.Zones[ns.zone].NS[name] != ip {
		logger.Fatal("%s has no matched recrods, %s.", alu.Caller, ip)
		return errors.New("No matched ip.")
	}

	delete(ns.config.Zones[ns.zone].NS, name)
	ns.save()

	return nil
}

// 更新 NS 記錄。
func (ns *NSService) Update(zone, name, ip string) error {
	ns.cs = NewConfigService()
	config, err := ns.cs.Read()
	if err != nil {
		return err
	}
	ns.config = config

	// Zone
	if _, ok := ns.config.Zones[zone]; !ok {
		logger.Fatal("%s has no matched zone name, %s.", alu.Caller(), zone)
		return errors.New("No matched zone name.")
	}
	ns.zone = zone

	// Name
	if _, ok := ns.config.Zones[ns.zone].NS[name]; !ok {
		logger.Fatal("%s no matched host name, %s.", alu.Caller(), name)
		return errors.New("No matched host name.")
	}

	ns.config.Zones[ns.zone].NS[name] = ip
	err = ns.save()
	if err != nil {
		return err
	}

	return nil
}

// 提交 NS 記錄到 PChome 網站。
func (ns *NSService) save() error {
	reader := strings.NewReader(ns.preparePostData().Encode())
	urlstr := ENDPOINT + "/dns_edit.php"
	req, err := http.NewRequest("POST", urlstr, reader)
	if err != nil {
		logger.Fatalf("%s creates http request failed, %s.", alu.Caller(), err.Error())
		return errors.New("Creating http request failed.")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	ns.Service.SetCookie(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("%s requesting failed, %s.", alu.Caller(), err.Error())
		return errors.New("Having http requesting failed.")
	}
	resp.Body.Close()
	if err := ns.cs.Save(&ns.config); err != nil {
		return err
	}

	return nil
}

// 準備提交的表單資料。
func (ns *NSService) preparePostData() url.Values {
	data := url.Values{}

	for i := 0; i < 5; i++ {
		data.Add("host_dn" + strconv.Itoa(i), "")
		data.Add("host_ip" + strconv.Itoa(i), "")
		data.Add("host_ipv6" + strconv.Itoa(i), "")
	}
	idx := 0
	for name, ip := range ns.config.Zones[ns.zone].NS {
		data.Set("host_dn" + strconv.Itoa(idx), name)
		data.Set("host_ip%d" + strconv.Itoa(idx), ip)
		idx++
	}

	for i := 0; i < 10; i++ {
		data.Add("subhostf%d" + strconv.Itoa(i), "")
		data.Add("contentf" + strconv.Itoa(i), "")
		data.Add("typef" + strconv.Itoa(i), "fwd")
		data.Add("fwd_titlef" + strconv.Itoa(i), "")
		data.Add("fwd_meta_tagf" + strconv.Itoa(i), "")
		data.Add("fwd_description_tagf" + strconv.Itoa(i), "")
	}

	data.Add("dn", ns.zone)
	data.Add("dns_mode", "1")

	return data
}

// 列舉 PChome 網站的 NS 記錄。
func (ns *NSService) List(zone string) (NS, error) {
	if len(zone) == 0 {
		logger.Fatalf("%s has empty zone name.", alu.Caller())
	}

	urlstr := "http://myname.pchome.com.tw/manage/dns_edit.htm?dn=" + zone
	req, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		logger.Fatalf("%s creates request failed, %s.", alu.Caller(), err.Error())
		return nil, errors.New("Cannot create a http request.")
	}
	ns.Service.SetCookie(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("%s requesting failed, %s.", alu.Caller(), err.Error())
		return nil, errors.New("Having http requesting failed.")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("%s reads http body failed, %s.", alu.Caller(), err.Error())
		return nil, errors.New("Reading http body failed.")
	}
	resp.Body.Close()
	slice, err := ns.parse(b)
	if err != nil {
		return nil, err
	}
	return slice, nil
}

// 解析 PChome DNSSEC 網頁。
func (ns *NSService) parse(raw []byte) (NS, error) {
	if len(raw) == 0 {
		logger.Printf("%s has empty raw.", alu.Caller())
		return nil, errors.New("Empty raw content to parse.")
	}

	reName := regexp.MustCompile(`host_dn\d" value="((?:\w+\.)+\w+)"`)
	names := reName.FindAllStringSubmatch(string(raw), -1)
	reIP := regexp.MustCompile(`host_ip\d" value="((?:\d{1,3}\.){3}\d{1,3})"`)
	ips := reIP.FindAllStringSubmatch(string(raw), -1)
	if len(names) != len(ips) {
		logger.Fatalf("%s has difference results.", alu.Caller())
		return nil, errors.New("NS data does not match regex patterns.")
	}

	record := make(NS)
	for i := 0; i < len(names); i++ {
		record[names[i][1]] = ips[i][1]
	}

	return record, nil
}
