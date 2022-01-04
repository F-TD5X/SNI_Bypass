package main

import (
	"flag"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/F-TD5X/SNI_Bypass/core"
	"github.com/F-TD5X/SNI_Bypass/core/certificate"
	"github.com/F-TD5X/SNI_Bypass/core/hosts"
	runtime "github.com/banzaicloud/logrus-runtime-formatter"
	log "github.com/sirupsen/logrus"
)

var (
	version       string
	isVerbose     = flag.Bool("v", false, "verbose mode")
	configPath    = flag.String("c", "./config.yml", "config file path")
	isShowVersion = flag.Bool("V", false, "current version of overture")
	enablePprof   = flag.Bool("pprof", false, "enable pprof")
	logPath       = flag.String("log", "./SNI_Bypass.log", "log file path")
	removeCA      = flag.Bool("remove-ca", false, "remove CA cert")
	trustCA       = flag.Bool("trust-ca", true, "trust CA cert")
)

func main() {
	flag.IntVar(&hosts.SetupHosts, "h", 0, "How to setup hosts: 0: system, 1: write to ./hosts.txt")
	flag.Parse()

	formatter := runtime.Formatter{ChildFormatter: &log.TextFormatter{
		FullTimestamp: true,
	}}
	formatter.Line = true
	log.SetFormatter(&formatter)

	if *enablePprof {
		go func() {
			log.Info(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	if *removeCA {
		certificate.RemoveCACert()
		os.Exit(0)
	}

	if *isVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if *logPath != "" {
		lf, err := os.OpenFile(*logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			log.Errorf("Unable to open log file: %s", *logPath)
		} else {
			log.SetOutput(io.MultiWriter(lf, os.Stdout))
		}
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	core.Serve(stop)
}
