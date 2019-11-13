package conf

import (
	"encoding/json"
	"fmt"
	"github.com/xvwvx/ipapk-server/utils"
	"io/ioutil"
)

var AppConfig *Config

type Config struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Proxy    string `json:"proxy"`
	Database string `json:"database"`
	IsUseAliyun bool `json:"isUseAliyun"`
	AccessID  string `json:"accessID"`
	AccessSecret  string `json:"accessSecret"`
	Bucket string `json:"bucket"`
	Visit string `json:"visit"`
}

func InitConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &AppConfig); err != nil {
		return err
	}
	return nil
}

func (c *Config) Addr() string {
	return fmt.Sprintf("%v:%v", c.Host, c.Port)
}

func (c *Config) ProxyURL() string {
	if c.Proxy == "" {
		localIp, err := utils.LocalIP()
		if err != nil {
			panic(err)
		}
		return fmt.Sprintf("https://%v:%v", localIp.String(), c.Port)
	}
	return c.Proxy
}

func (c *Config) VisitURL() string {
	if c.IsUseAliyun {
		return c.Visit
	}
	return c.ProxyURL()
}