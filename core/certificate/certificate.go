package certificate

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
)

var (
	certFileName      = "ca.crt"
	keyFileName       = "ca.key"
	ca_CN             = "SNI Root CA"
	ca                tls.Certificate
	leaf              *x509.Certificate
	certCache         sync.Map
	serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)
	privKey           *rsa.PrivateKey
)

func GetCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	commonName, err := publicsuffix.EffectiveTLDPlusOne(info.ServerName)
	if err != nil {
		log.Warning("Can't not parse common name from client Hello. ", err)
		return nil, err
	}
	if cert, ok := certCache.Load(commonName); ok {
		return cert.(*tls.Certificate), nil
	}

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Warning("Can't generate serial Number. ", err)
		return nil, err
	}
	now := time.Now()
	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&x509.Certificate{
			SerialNumber: serialNumber,
			NotBefore:    now,
			NotAfter:     now.AddDate(5, 0, 0),
			Subject: pkix.Name{
				CommonName:   commonName,
				Organization: []string{commonName},
			},
			DNSNames: []string{"*." + commonName, commonName},
		},
		leaf,
		privKey.Public(),
		ca.PrivateKey,
	)
	if err != nil {
		log.Warning("Can't generate server cert. ", err)
		return nil, err
	}
	cert := &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  privKey,
	}
	certCache.Store(commonName, cert)
	return cert, nil
}

func GenCA() tls.Certificate {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Country:      []string{"CN"},
			Organization: []string{"SNI"},
			CommonName:   ca_CN,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privKey.PublicKey, privKey)
	if err != nil {
		log.Fatal("Can't generate CA cert. ", err)
		return tls.Certificate{}
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privKey)})
	ioutil.WriteFile(certFileName, certPEM, 0777)
	ioutil.WriteFile(keyFileName, keyPEM, 0777)
	cert, _ := tls.X509KeyPair(certPEM, keyPEM)

	return cert
}

func init() {
	if _, err := os.Stat(certFileName); err == nil {
		ca, err = tls.LoadX509KeyPair(certFileName, keyFileName)
		if err != nil {
			log.Error("Can't open CA files. ", err)
			return
		}
		leaf, err = x509.ParseCertificate(ca.Certificate[0])
		if err != nil {
			log.Error("Wrong CA files. ", err)
			return
		}
		privKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	} else {
		ca = GenCA()
		leaf, _ = x509.ParseCertificate(ca.Certificate[0])
		privKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	}
	if !VerifyCert() {
		err := TrustCACert()
		if err != nil {
			log.Error(err)
			os.Exit(-1)
		}
	}
}

func TrustCACert() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("certutil", "-user", "-addstore", "Root", certFileName)
	case "linux":
	case "darwin":
	}
	out, err := cmd.Output()
	s := string(out)
	c := cmd.String()
	log.Debug(c, s)
	if err != nil {
		log.Error("Can't trust CA cert. ", err)
		return err
	}
	return nil
}

func RemoveCACert() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("certutil", "-user", "-delstore", "Root", ca_CN)
	case "linux":
	case "darwin":
	}
	err := cmd.Run()
	if err != nil {
		log.Error("Can't remove CA cert. ", err)
		return err
	}
	return nil
}

func VerifyCert() bool {
	cmd := exec.Command("certutil", "-user", "-verifystore", "Root", ca_CN)
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	if strings.Contains(string(out), "NTE_NOT_FOUND") {
		return false
	}
	return true
}
