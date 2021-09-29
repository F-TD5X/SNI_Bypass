package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func listen() {
	go func() {
		log.Warn(http.ListenAndServe(":80", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host, http.StatusMovedPermanently)
			}),
		))
	}()
	listenTLS()
}
func main() {
	//go http.ListenAndServe("0.0.0.0:6060", nil)
	c := make(chan os.Signal, 10)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGSEGV)
	setHosts()
	go listen()
	s := <-c
	log.Info("Exiting... ", s)
	restoreHosts()
}
