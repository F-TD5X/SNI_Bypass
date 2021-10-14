package main

import (
	"crypto/tls"
	"io"
	"sync"

	log "github.com/sirupsen/logrus"
)

func ServeTLS() {
	l, err := tls.Listen("tcp", ":https", &tls.Config{GetCertificate: getCertificate})
	if err != nil {
		log.Error("Can't start tls server: ", err)
		return
	}
	for {
		if conn, err := l.Accept(); err != nil {
			log.Warn("Lose tls conn: ", err)
		} else {
			go forward(conn.(*tls.Conn))
		}
	}
}

func forward(s *tls.Conn) {
	var err error
	if err := s.Handshake(); err != nil {
		//log.Warningf("Handshake with client failed. With Servername: %s. Error: %s", s.ConnectionState().ServerName, err)
		return
	}
	if _, ok := c.Hosts[s.ConnectionState().ServerName]; !ok {
		return
	}
	var i *tls.Conn
	for _, upstream := range getUpstreams(s.ConnectionState().ServerName) {
		log.Debugf("Trying upstream %s for %s.", upstream, s.ConnectionState().ServerName)
		i, err = tls.Dial("tcp", upstream+":443", &tls.Config{InsecureSkipVerify: true, ServerName: getSNI(s.ConnectionState().ServerName)})
		if err != nil {
			log.Errorf("Dial TLS failed to %s. %s", upstream, err)
		} else {
			log.Infof("Use upstream %s for %s.", upstream, s.ConnectionState().ServerName)
			break
		}
	}
	if i == nil {
		log.Warningf("No upstream connectable for host: %s", s.ConnectionState().ServerName)
		return
	}
	var g sync.WaitGroup
	g.Add(2)
	go transfer(i, s, &g)
	go transfer(s, i, &g)
	g.Wait()

	if err := i.Close(); err != nil {
		log.Warning("Can't stop object conn. ", err)
	}
	if err := s.Close(); err != nil {
		log.Warning("Cant' stop client conn. ", err)
	}
}

func transfer(src io.Reader, dst io.Writer, g *sync.WaitGroup) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Warning("IO transfer error.")
	}
	if err := dst.(*tls.Conn).CloseWrite(); err != nil {
		log.Warning("Can't stop object writing. ", err)
	}
	if err := src.(*tls.Conn).CloseWrite(); err != nil {
		log.Warning("Can't stop client writing. ", err)
	}
	g.Done()
}
