package cert

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

const (
	caValidityDuration = 3 * 365 * 24 * time.Hour
	caKeyUsage         = x509.KeyUsageDigitalSignature |
		x509.KeyUsageContentCommitment |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement |
		x509.KeyUsageCertSign |
		x509.KeyUsageCRLSign
)

type CA struct {
	*Certificate
}

func (c CA) GenerateCert(names []string) (*Certificate, error) {
	return generateCert(c.Certificate, names)
}

func generateCA(name string) (CA, error) {
	validityStart := time.Now().UTC()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: name},
		NotBefore:             validityStart,
		NotAfter:              validityStart.Add(caValidityDuration),
		KeyUsage:              caKeyUsage,
		BasicConstraintsValid: true,
		IsCA:               true,
		MaxPathLen:         0,
		MaxPathLenZero:     true,
		SignatureAlgorithm: x509.ECDSAWithSHA256,
	}
	key, err := generateCAKey()
	if err == nil {
		cert, err := createCertificate(template, template, key, key)
		return CA{Certificate: cert}, err
	}
	return CA{}, err
}

func LoadCA(certFile, keyFile, name string) (CA, error) {
	if tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile); err == nil {
		if key, ok := tlsCert.PrivateKey.(*ecdsa.PrivateKey); ok {
			tlsCert.PrivateKey = Key{PrivateKey: key}
			tlsCert.Leaf, err = x509.ParseCertificate(
				tlsCert.Certificate[0],
			)
			return CA{Certificate: (*Certificate)(&tlsCert)}, err
		}
	}
	ca, err := generateCA(name)
	if err == nil {
		err = ca.WritePEMFiles(certFile, keyFile)
		return ca, err
	}
	return CA{}, err
}
