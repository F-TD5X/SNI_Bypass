package certificate_test

import (
	"testing"

	"github.com/F-TD5X/SNI_Bypass/core/certificate"
)

/*
Because the program trust a Root CA, so we can't test the CA cert operation automatically.
This test has to be run manually.
*/
func TestCAOps(t *testing.T) {

	err := certificate.TrustCACert()
	if err != nil {
		t.Error(err)
	}
	if !certificate.VerifyCert() {
		t.Fatal("Trust CA cert failed")
	}
	err = certificate.RemoveCACert()
	if err != nil {
		t.Error(err)
	}
	if certificate.VerifyCert() {
		t.Fatal("Remove CA cert failed")
	}
}
