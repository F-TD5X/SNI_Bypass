package main

import (
	"fmt"
	"io/ioutil"
	"runtime"

	log "github.com/sirupsen/logrus"
)

var (
	origHosts string
	Hosts     string
	Hostsfile string
)

func setHosts() {
	switch runtime.GOOS {
	case "darwin":
		Hostsfile = "/etc/hosts"
	case "windows":
		Hostsfile = "C:/Windows/System32/drivers/etc/hosts"
	case "linux":
		Hostsfile = "/etc/hosts"
	}
	buf, err := ioutil.ReadFile(Hostsfile)
	if err != nil {
		log.Error("Can't read hosts file. ", err)
		return
	}
	origHosts = string(buf)
	for _, v := range hosts {
		Hosts += fmt.Sprintf("127.0.0.1 %s #SNI \n", v)
	}
	err = ioutil.WriteFile(Hostsfile, []byte(origHosts+"\n\n"+Hosts), 0)
	if err != nil {
		log.Error("Can't write hosts file. ", err)
		return
	}
}

func restoreHosts() {
	err := ioutil.WriteFile(Hostsfile, []byte(origHosts), 0)
	if err != nil {
		log.Error("Can't restore hosts file. ", err)
		return
	}
}
