package core

import (
	"net/http"
	"os"

	"github.com/F-TD5X/SNI_Bypass/core/hosts"
	"github.com/F-TD5X/SNI_Bypass/core/tls"

	log "github.com/sirupsen/logrus"
)

func Serve(stop chan os.Signal) {
	err := hosts.Setup()
	defer func() {
		err := hosts.Restore()
		if err != nil {
			log.Errorf("Restore hosts file failed: %s", err)
		}
	}()

	if err != nil {
		log.Error("Can't setup hosts file. ", err)
		return
	}

	go func() {
		log.Warn(http.ListenAndServe(":80", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+r.Host, http.StatusMovedPermanently)
			}),
		))
	}()
	go tls.ServeTLS()

	<-stop
}
