package main

import (
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
	Entrys []struct {
		Hosts    []string `yaml:"hosts"`
		Enable   bool     `yaml:"enable"`
		Group    string   `yaml:"group"`
		Upstream []string `yaml:"upstream"`
		Sni      string   `yaml:"SNI"`
	} `yaml:"Entry"`
	HostGroups map[string][]string `yaml:"UpstreamGroups"`
	Hosts      map[string]HostItem
}

func getUpstreams(host string) []string {
	if v, ok := c.Hosts[host]; ok {
		if g, ok := c.HostGroups[v.Group]; ok {
			return g
		} else {
			return v.Upstream
		}
	} else {
		return nil
	}
}

var (
	c     YAML
	hosts []string
)

func getSNI(host string) string {
	if v, ok := c.Hosts[host]; ok {
		if v.Sni != "" {
			return v.Sni
		}
	}
	return "baidu.com"
}

func init() {
	data, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Error(err)
		return
	}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Error(err)
		return
	}
	c.Hosts = make(map[string]HostItem)
	for _, v := range c.Entrys {
		if v.Enable {
			hosts = append(hosts, v.Hosts...)
			for _, h := range v.Hosts {
				c.Hosts[h] = HostItem{Group: v.Group, Upstream: v.Upstream, Sni: v.Sni}
			}
		}
	}
}
