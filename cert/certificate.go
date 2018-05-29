package cert

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	mrand "math/rand"
	"time"
)

const (
	certValidityDuration = 24 * time.Hour
	certKeyUsage         = x509.KeyUsageDigitalSignature |
		x509.KeyUsageContentCommitment |
		x509.KeyUsageKeyEncipherment |
		x509.KeyUsageDataEncipherment |
		x509.KeyUsageKeyAgreement
)

var (
	errInvalidCACert = errors.New("invalid CA certificate")
	errInvalidKey    = errors.New("invalid key")
)

type Certificate tls.Certificate

func (c *Certificate) DER() []byte {
	if len(c.Certificate) > 0 {
		return c.Certificate[0]
	}
	return nil
}

func (c *Certificate) PEM() []byte {
	return pem.EncodeToMemory(
		&pem.Block{Type: "CERTIFICATE", Bytes: c.DER()},
	)
}

func (c *Certificate) TLS() *tls.Certificate {
	return (*tls.Certificate)(c)
}

func (c *Certificate) WritePEMFiles(certFile, keyFile string) error {
	if key, ok := c.PrivateKey.(Key); ok {
		err := ioutil.WriteFile(keyFile, key.PEM(), 0600)
		if err == nil {
			err = ioutil.WriteFile(certFile, c.PEM(), 0644)
		}
		return err
	}
	return errInvalidKey
}

func randomSerialNumber() *big.Int {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, max)
	if err == nil {
		return serialNumber
	}
	return new(big.Int).Mul(
		big.NewInt(mrand.Int63()),
		big.NewInt(mrand.Int63()),
	)
}

func createCertificate(
	template, parent *x509.Certificate,
	certKey, signKey Key,
) (*Certificate, error) {
	certDER, err := x509.CreateCertificate(
		rand.Reader, template, parent, certKey.Public(), signKey,
	)
	if err == nil {
		leaf, err := x509.ParseCertificate(certDER)
		if err == nil {
			cert := &Certificate{
				Certificate: [][]byte{certDER},
				PrivateKey:  certKey,
				Leaf:        leaf,
			}
			return cert, nil
		}
	}
	return nil, err
}

func generateCert(ca *Certificate, names []string) (*Certificate, error) {
	if !ca.Leaf.IsCA {
		return nil, errInvalidCACert
	}
	validityStart := time.Now().Add(-1 * time.Hour).UTC()
	template := &x509.Certificate{
		SerialNumber:          randomSerialNumber(),
		Subject:               pkix.Name{CommonName: names[0]},
		NotBefore:             validityStart,
		NotAfter:              validityStart.Add(certValidityDuration),
		KeyUsage:              certKeyUsage,
		BasicConstraintsValid: true,
		DNSNames:              names,
		SignatureAlgorithm:    x509.ECDSAWithSHA256,
	}
	key, err := generateCertKey()
	if err == nil {
		if caKey, ok := ca.PrivateKey.(Key); ok {
			return createCertificate(template, ca.Leaf, key, caKey)
		}
		err = errInvalidKey
	}
	return nil, err
}
