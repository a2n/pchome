package pchome

import (
	"net/http"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"net/url"
	"errors"

	"github.com/a2n/alu"
)

type DNSSEC struct {
	KeyTag uint16
	Algorithm uint8
	Digest string
}

type DNSSECService struct {
	Service *Service
	cs *ConfigService
	config Config
	zone string
}

func (ds *DNSSECService) Add(zone string, keyTag uint16, algorithm uint8, digest string) error {
	ds.cs = NewConfigService()
	config, err := ds.cs.Read()
	if err != nil {
		return err
	}
	ds.config = config

	// Zone
	if _, ok := ds.config.Zones[zone]; !ok {
		logger.Fatal("%s has no such zone name, %s.", alu.Caller(), zone)
		return errors.New("No such zone name")
	}
	zoneObj := ds.config.Zones[zone]
	ds.zone = zone

	// Max records count.
	if len(zoneObj.DNSSEC) == 5 {
		logger.Fatal("%s, the zone(%s) has reached the max DNESEC records count 5.", alu.Caller(), zone)
		return errors.New("The DNSSEC records of this zone is reaching the max count 5, delete some records first.\n")
	}

	// Find existed records.
	for _, dnssec := range zoneObj.DNSSEC {
		if dnssec.KeyTag == keyTag && dnssec.Algorithm == algorithm && dnssec.Digest == digest {
			logger.Fatal("%s has duplicated.")
			return errors.New("Duplicated record.")
		}
	}

	record := DNSSEC {
		KeyTag: keyTag,
		Algorithm: algorithm,
		Digest: digest,
	}
	zoneObj.DNSSEC = append(zoneObj.DNSSEC, record)
	ds.config.Zones[ds.zone] = zoneObj
	ds.save()

	return nil
}

func (ds *DNSSECService) Delete(zone string, keyTag uint16, algorithm uint8, digest string) error {
	ds.cs = NewConfigService()
	config, err := ds.cs.Read()
	if err != nil {
		return err
	}
	ds.config = config

	// Zone
	if _, ok := ds.config.Zones[zone]; !ok {
		logger.Fatal("%s has no matched zone name, %s.", alu.Caller(), zone)
		return errors.New("No matched zone name.")
	}
	zoneObj := ds.config.Zones[zone]
	ds.zone = zone

	// Find existed records.
	found := false
	for i, dnssec := range zoneObj.DNSSEC {
		if dnssec.KeyTag == keyTag && dnssec.Algorithm == algorithm && dnssec.Digest == digest {
			found = true
			zoneObj.DNSSEC = append(zoneObj.DNSSEC[:i], zoneObj.DNSSEC[i + 1:]...)
			ds.config.Zones[zone] = zoneObj
			ds.save()
			break
		}
	}
	if !found {
		logger.Fatalf("%s has no matched DNSSEC record.", alu.Caller())
		return errors.New("No matched DNSSEC record.")
	}

	return nil
}

func (ds *DNSSECService) save() error {
	reader := strings.NewReader(ds.preparePostData().Encode())
	urlstr := ENDPOINT + "/set_dnssec.php"
	req, err := http.NewRequest("POST", urlstr, reader)
	if err != nil {
		logger.Fatalf("%s creates http request failed, %s.", alu.Caller(), err.Error())
		return errors.New("Creating http request failed.")
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	ds.Service.SetCookie(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("%s requesting failed, %s.", alu.Caller(), err.Error())
		return errors.New("Having http requesting failed.")
	}
	resp.Body.Close()
	ds.cs.Save(&ds.config)

	return nil
}

func (ds *DNSSECService) preparePostData() url.Values {
	data := url.Values{}

	for i := 0; i < 5; i++ {
		data.Add("KeyTag" + strconv.Itoa(i), "")
		data.Add("alg" + strconv.Itoa(i), "")
		data.Add("DS" + strconv.Itoa(i), "")
	}

	for i, dnssec := range ds.config.Zones[ds.zone].DNSSEC {
		data.Set("KeyTag" + strconv.Itoa(i), strconv.Itoa(int(dnssec.KeyTag)))
		data.Set("alg" + strconv.Itoa(i), strconv.Itoa(int(dnssec.Algorithm)))
		data.Set("DS" + strconv.Itoa(i), dnssec.Digest)
	}

	data.Add("dn", ds.zone)

	return data
}

func (ds *DNSSECService) List(zone string) ([]DNSSEC, error) {
	if len(zone) == 0 {
		logger.Fatalf("%s has empty zone name.", alu.Caller())
	}

	urlstr := "http://myname.pchome.com.tw/manage/set_dnssec.htm?dn=" + zone
	req, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		logger.Fatalf("%s creates http request failed, %s.", alu.Caller(), err.Error())
		return nil, errors.New("Cannot create a http request.")
	}
	ds.Service.SetCookie(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("%s requesting failed, %s.", alu.Caller(), err.Error())
		return nil, errors.New("Having http requesting failed.")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("%s reads http body failed, %s.", alu.Caller(), err.Error())
		return nil, errors.New("Reads http body failed.")
	}
	resp.Body.Close()

	slice, err := ds.parse(b)
	if err != nil {
		return nil, err
	}
	return slice, nil
}

func (ds *DNSSECService) parse(raw []byte) ([]DNSSEC, error) {
	if len(raw) == 0 {
		logger.Printf("%s has empty raw.", alu.Caller())
		return nil, errors.New("Empty content to parse.")
	}

	reKeyTag := regexp.MustCompile(`KeyTag\d" value="(\d{1,6})`)
	keyTags := reKeyTag.FindAllStringSubmatch(string(raw), -1)
	reAlgorithm := regexp.MustCompile(`alg\d" value="(\d{1,3})`)
	algorithms := reAlgorithm.FindAllStringSubmatch(string(raw), -1)
	reDigest := regexp.MustCompile(`DS\d" value="([\d|\w]+)`)
	digests := reDigest.FindAllStringSubmatch(string(raw), -1)

	if len(keyTags) != len(algorithms) && len(algorithms) != len(digests) {
		logger.Fatalf("%s has difference results.", alu.Caller())
		return nil, errors.New("DNSSEC data does not match regex patterns.")
	}

	records := make([]DNSSEC, 0)
	for i := 0; i < len(keyTags); i++ {
		keyTag, err := strconv.Atoi(keyTags[i][1])
		if err != nil {
			logger.Printf("%s parse string(%s) failed, %s.", alu.Caller(), keyTags[i][1], err.Error())
			return nil, errors.New("Parse key tag failed.")
		}

		algorithm, err := strconv.Atoi(algorithms[i][1])
		if err != nil {
			logger.Printf("%s parse string(%s) failed, %s.", alu.Caller(), algorithms[i][1], err.Error())
			return nil, errors.New("Parse algorithm failed.")
		}

		records = append(records, DNSSEC {
			KeyTag: uint16(keyTag),
			Algorithm: uint8(algorithm),
			Digest: digests[i][1],
		})
	}

	return records, nil
}
