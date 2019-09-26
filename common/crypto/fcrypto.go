package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"strings"
)

func DecodeBytes2ECDSAPublicKey(pemBytes []byte) (*ecdsa.PublicKey, error) {
	p, _ := pem.Decode(pemBytes)
	if p == nil {
		return nil, errors.New("cert is empty1")
	}
	certificate, err := x509.ParseCertificates(p.Bytes)
	if err != nil {
		return nil, err
	}
	if len(certificate) == 0 {
		return nil, errors.New("cert is empty2")
	}
	cert := certificate[0]
	if !strings.EqualFold(cert.PublicKeyAlgorithm.String(), x509.ECDSA.String()) {
		return nil, errors.New("cert type error:" + cert.PublicKeyAlgorithm.String() + ",need type:" + x509.ECDSA.String())
	}
	publicKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("parse public key error")
	}
	return publicKey, nil
}
