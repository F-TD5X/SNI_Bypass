package hosts

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/F-TD5X/SNI_Bypass/core/config"
)

var (
	SetupHosts int
	flag       = "# SNI"
)

func getHostsFilePath() string {
	if SetupHosts == 1 {
		return "hosts.txt"
	} else {
		switch runtime.GOOS {
		case "darwin":
			return "/etc/hosts"
		case "windows":
			return "C:/Windows/System32/drivers/etc/hosts"
		case "linux":
			return "/etc/hosts"
		}
		return ""
	}
}

func Setup() error {
	if SetupHosts == 0 {
		buf, err := ioutil.ReadFile(getHostsFilePath())
		if err != nil {
			return err
		}
		originalHosts := string(buf)
		err = ioutil.WriteFile(getHostsFilePath(), []byte(originalHosts+"\n"+config.GenHostsString()), 0644)
		if err != nil {
			return err
		}
	} else if SetupHosts == 1 {
		err := ioutil.WriteFile(getHostsFilePath(), []byte(config.GenHostsString()), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func Restore() error {
	if SetupHosts == 0 {
		buf, err := ioutil.ReadFile(getHostsFilePath())
		if err != nil {
			return err
		}
		hosts := strings.Split(string(buf), "\n")
		f, err := os.OpenFile(getHostsFilePath(), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		for _, v := range hosts {
			if strings.Contains(v, flag) {
				continue
			} else {
				fmt.Fprintf(f, "%s\n", v)
			}
		}
	}
	return nil
}
