package config

import (
	"fmt"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

type HostItem struct {
	Group    string
	Upstream []string
	Sni      string
}

type YAML struct {
	DefaultSNI string `yaml:"DefaultSNI"`
	Entrys     []struct {
		Name     string   `yaml:"name"`
		Hosts    []string `yaml:"hosts"`
		Enable   bool     `yaml:"enable"`
		Group    string   `yaml:"group"`
		Upstream []string `yaml:"upstream"`
		Sni      string   `yaml:"SNI"`
	} `yaml:"Entry"`
	HostGroups map[string][]string `yaml:"UpstreamGroups"`
	Hosts      map[string]HostItem
}

func GetUpstreams(host string) []string {
	if v, ok := Config.Hosts[host]; ok {
		if g, ok := Config.HostGroups[v.Group]; ok {
			return g
		} else {
			return v.Upstream
		}
	} else {
		return nil
	}
}

var (
	Config YAML
	Hosts  []string
)

func GetSNI(host string) string {
	if v, ok := Config.Hosts[host]; ok {
		if v.Sni != "" {
			return v.Sni
		}
	}
	return Config.DefaultSNI
}

func init() {
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Error("Can't open config file: ", err)
		return
	}
	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		log.Error("Can't pharse config file: ", err)
		return
	}
	Config.Hosts = make(map[string]HostItem)
	for _, v := range Config.Entrys {
		if v.Enable {
			Hosts = append(Hosts, v.Hosts...)
			for _, h := range v.Hosts {
				Config.Hosts[h] = HostItem{Group: v.Group, Upstream: v.Upstream, Sni: v.Sni}
			}
		}
	}
}

func GenHostsString() string {
	ret := ""
	for _, v := range Hosts {
		ret += fmt.Sprintf("127.0.0.1 %s # SNI \n", v)
	}
	return ret
}
