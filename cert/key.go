package cert

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
)

type Key struct {
	*ecdsa.PrivateKey
}

func (k Key) DER() []byte {
	keyDER, err := x509.MarshalECPrivateKey(k.PrivateKey)
	if err == nil {
		return keyDER
	}
	return nil
}

func (k Key) PEM() []byte {
	return pem.EncodeToMemory(
		&pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: k.DER()},
	)
}

func generateKey(c elliptic.Curve) (Key, error) {
	key, err := ecdsa.GenerateKey(c, rand.Reader)
	return Key{PrivateKey: key}, err
}

func generateCAKey() (Key, error) {
	return generateKey(elliptic.P384())
}

func generateCertKey() (Key, error) {
	return generateKey(elliptic.P256())
}
