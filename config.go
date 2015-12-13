package pchome
/*
	指令操作 PCHome 買網址	
 */

import (
	"time"
	"fmt"
	"encoding/json"
	"os"
	"io/ioutil"
	"sort"
	"net/url"
	"net/http"
	"errors"

	"github.com/a2n/alu"
)

// 預設的組態檔案位置
const DefaultConfigPath = ".pchome"

// 組態服務結構
type ConfigService struct {
	Service *Service
}

// 取得組態服務。
func NewConfigService() *ConfigService {
	logger = alu.NewLogger("log")
	return &ConfigService{}
}

// 初始組態服務
func (cs *ConfigService) Init() error {
	b, err := ioutil.ReadFile(DefaultConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := cs.initNew(); err != nil {
				return err
			}
		} else {
			logger.Fatalf("%s read config file failed, %s.", alu.Caller(), err.Error())
		}
	}

	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		logger.Fatalf("%s unmarshal json failed, %s.", alu.Caller(), err.Error())
	}

	return nil
}

// 初始全新組態。
func (cs *ConfigService) initNew() error {
	config := Config {
		UpdatedAt: time.Now().Unix(),
		Zones: make(map[string]Zone),
	}

	// Basic info 
	fmt.Print("Paste your email here: ")
	_, err := fmt.Scanln(&config.Email)
	if err != nil {
		logger.Fatalf("%s scan email string failed, %s.", alu.Caller(), err.Error())
		return errors.New("Scan email string failed.")
	}
	if len(config.Email) == 0 {
		logger.Fatal("%s has empty email.", alu.Caller())
		return errors.New("Empty email.")
	}

	fmt.Print("Paste your password here: ")
	_, err = fmt.Scanln(&config.Password)
	if err != nil {
		logger.Fatalf("%s scan password string failed, %s.", alu.Caller(), err.Error())
		return errors.New("Scan password string failed.")
	}
	if len(config.Password) == 0 {
		logger.Fatal("%s has empty password.", alu.Caller())
		return errors.New("Empty password.")
	}

	key, err := cs.DoGetKey(config.Email, config.Password)
	if err != nil {
		return err
	}

	if len(key) == 0 {
		logger.Fatal("%s gets a empty key.", alu.Caller())
		return errors.New("Your email or password is wrong.")
	}

	// Zones & Records
	zones, err := cs.UpdateZones(NewService(key))
	if err != nil {
		return err
	} else {
		config.Zones = zones
		if err = cs.Save(&config); err != nil {
			return err
		}
	}

	return nil
}

// 取得 PCHome 存取鑰匙
func (cs *ConfigService) GetKey() (string, error) {
	config, err := cs.Read()
	if err != nil {
		return "", err
	}

	key, err := cs.DoGetKey(config.Email, config.Password)
	if err != nil {
		return "", err
	}

	return key, nil
}

// 從網站取得 PCHome 存取鑰匙
func (cs *ConfigService) DoGetKey(email, password string) (string, error) {
	if len(email) == 0 {
		logger.Printf("%s has empty email.", alu.Caller())
		return "", errors.New("Empty email.")
	}

	if len(password) == 0 {
		logger.Printf("%s has empty password.", alu.Caller())
		return "", errors.New("Empty password.")
	}

	urlstr := "https://login.pchome.com.tw/adm/person_sell.htm"
	data := url.Values {
		"mbrid": []string{email},
		"mbrpass": []string{password},
		"chan": []string{"P000007"},
		"ltype": []string{"checklogin"},
	}

	resp, err := http.DefaultClient.PostForm(urlstr, data)
	if err != nil {
		logger.Printf("%s http requesting failed, %s.", alu.Caller(), err.Error())
		return "", errors.New("Http requesting failed.")
	}

	key := ""
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "loginkuser" {
			key = cookie.Value
			break
		}
	}

	resp.Body.Close()

	return key, nil
}

// 更新組態內容。
func (cs *ConfigService) Update() error {
	// Open
	config, err := cs.Read()
	if err != nil {
		return err
	}
	config.UpdatedAt = time.Now().Unix()

	// Zones & Records
	key, err := cs.DoGetKey(config.Email, config.Password)
	if err != nil {
		return err
	}

	zones, err := cs.UpdateZones(NewService(key))
	if err != nil {
		return err
	} else {
		config.Zones = zones
		if err = cs.Save(&config); err != nil {
			return err
		}
	}

	return nil
}

// 更新 zone 內容。
func (cs *ConfigService) UpdateZones(s *Service) (map[string]Zone, error) {
	zones := s.NewZoneService().List().Do()
	keys := make([]string, 0)
	for k, _ := range zones {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ns := s.NewNSService()
	ds := s.NewDNSSECService()
	zone := make(map[string]Zone)
	for k, v := range keys {
		logger.Printf("%s received %s dns records. %d/%d) ", alu.Caller(), v, k + 1, len(zones))
		fmt.Printf("%d/%d) Received %s dns records.\n", k + 1, len(zones), v)

		nsSlice, err := ns.List(v)
		if err != nil {
			return nil, err
		}

		dnssecSlice, err := ds.List(v)
		if err != nil {
			return nil, err
		}

		zone[v] = Zone {
			NS: nsSlice,
			DNSSEC: dnssecSlice,
		}
	}

	return zone, nil
}

// 讀取本地組態。
func (cs *ConfigService) Read() (Config, error) {
	b, err := ioutil.ReadFile(DefaultConfigPath)
	if err != nil {
		logger.Fatalf("%s read configuration file failed, %s.", alu.Caller(), err.Error())
		return Config{}, errors.New("Read configuration file failed.")
	}

	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		logger.Fatalf("%s unmarshal json failed, %s.", alu.Caller(), err.Error())
		return config, errors.New("Unmarshal configuration json failed.")
	}

	return config, nil
}

// 移除組態檔案。
func (cs *ConfigService) Remove() error {
	err := os.Remove(DefaultConfigPath)
	if err != nil {
		logger.Fatalf("%s remove the configuration file failed, %s.", alu.Caller(), err.Error())
		return errors.New("Failed to remove the configuration file.")
	}

	logger.Printf("%s remove the configuration file successfully", alu.Caller())
	return nil
}

// 儲存組態內容。
func (cs *ConfigService) Save(config *Config) error {
	if config == nil {
		logger.Printf("%s has nil config.", alu.Caller())
		return errors.New("nil config.")
	}

	// Write
	b, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		logger.Fatalf("%s marshal json failed, %s.", alu.Caller(), err.Error())
		return errors.New("Marshal json failed.")
	}

	file, err := os.OpenFile(DefaultConfigPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			file, err = os.Create(DefaultConfigPath)
			if err != nil {
				logger.Fatalf("%s creates configuration file failed, %s.", alu.Caller(), err.Error())
				return errors.New("Create configuration file failed.")
			} else {
				file, err = os.OpenFile(DefaultConfigPath, os.O_RDWR, os.ModePerm)
				if err != nil {
					logger.Fatalf("%s opens configuration file failed, %s.", alu.Caller(), err.Error())
					return errors.New("Open configuration file failed.")
				}
			}
		} else {
			logger.Fatalf("%s opens configuration file failed, %s.", alu.Caller(), err.Error())
			return errors.New("Open configuration file failed.")
		}
	}

	_, err = file.Write(b)
	if err != nil {
		file.Close()
		logger.Fatalf("%s write configuration file failed, %s.", alu.Caller(), err.Error())
		return errors.New("Writing configuration file failed.")
	}

	logger.Printf("%s write configuration file successfully.", alu.Caller())
	file.Close()
	return nil
}

// 登出 PChome 網站。
func (cs *ConfigService) Logout() error {
	if len(cs.Service.Key) == 0 {
		return errors.New("Empty access token.")
	}

	urlstr := "https://login.pchome.com.tw/adm/logout.php"
	req, err := http.NewRequest("GET", urlstr, nil)
	if err != nil {
		logger.Printf("%s creates http request failed, %s.", alu.Caller(), err.Error())
		return errors.New("Cannot create a http request.")
	}
	cs.Service.SetCookie(req)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		logger.Printf("%s requesting failed, %s.", alu.Caller(), err.Error())
		return errors.New("Cannot create a http request.")
	}
	return nil
}

// 組態結構，記錄 Email、密碼、Zones 和最後更新時間。
type Config struct {
	Email string
	Password string
	Zones map[string]Zone
	UpdatedAt int64
}
