package main

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"os"
	"time"

	"github.com/ooqls/getset/crypto/keys"
)

func main() {
	ca, err := keys.CreateX509CA(keys.WithTemplate(x509.Certificate{
		Subject:   pkix.Name{CommonName: "Go Auth CA #1"},
		Issuer:    pkix.Name{CommonName: "Go Auth CA #1"},
		NotBefore: time.Now().Add(-2 * 24 * time.Hour),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
		IsCA:      true,
		KeyUsage:  x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}))
	if err != nil {
		panic(err)
	}

	caKeyPem, caCertPem := ca.Pem()
	os.WriteFile("ca/ca.key", caKeyPem, 0644)
	os.WriteFile("ca/ca.crt", caCertPem, 0644)

	// caCertPem, err := os.ReadFile("ca/ca.crt")
	// if err != nil {
	// 	panic(err)
	// }
	// caKeyPem, err := os.ReadFile("ca/ca.key")
	// if err != nil {
	// 	panic(err)
	// }

	err = keys.InitCA(caKeyPem, caCertPem)
	if err != nil {
		panic(err)
	}

	// ca := keys.CA()
	certs, err := keys.CreateX509(*ca,
		keys.WithCommonName("verylocal"),
		keys.WithDNSNames([]string{"verylocal", "localhost"}),
		keys.WithExtKeyUsage([]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}),
		keys.WithNotBefore(time.Now().Add(-2*24*time.Hour)),
		keys.WithNotAfter(time.Now().Add(90*24*time.Hour)),
	)
	if err != nil {
		panic(err)
	}

	certKeyPem, certCertPem := certs.Pem()
	os.WriteFile("server.key", certKeyPem, 0644)
	os.WriteFile("server.crt", certCertPem, 0644)

	jwtKey, err := keys.NewRSA()
	if err != nil {
		panic(err)
	}

	jwtKeyPem, pubkey := jwtKey.Pem()
	os.WriteFile("jwt.key", jwtKeyPem, 0644)
	os.WriteFile("jwt.pub", pubkey, 0644)
}
