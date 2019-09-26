package crypto

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"

	"github.com/cloudflare/cfssl/csr"
)

func GenCertificateRequest(name string, priKey *ecdsa.PrivateKey) ([]byte, error) {
	cr := csr.CertificateRequest{}
	bkr := csr.NewKeyRequest()
	cr.KeyRequest = &csr.KeyRequest{A: bkr.A, S: bkr.S}
	cr.CN = name
	hostname, _ := os.Hostname()
	if hostname != "" {
		cr.Hosts = make([]string, 1)
		cr.Hosts[0] = hostname
	}
	return csr.Generate(priKey, &cr)
}

func ParsePubKeyFromCert(cert []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(cert)

	if block.Type == "NEW CERTIFICATE REQUEST" || block.Type == "CERTIFICATE REQUEST" {
		csrReq, err := x509.ParseCertificateRequest(block.Bytes)
		if err != nil {
			return nil, err
		}
		lowLevelKey, ok := csrReq.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("invalid raw material. Expected *ecdsa.PublicKey")
		}
		return lowLevelKey, nil
	} else if block.Type == "CERTIFICATE" {
		x509Cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		lowLevelKey, ok := x509Cert.PublicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, errors.New("invalid raw material. Expected *ecdsa.PublicKey")
		}
		return lowLevelKey, nil
	}
	return nil, errors.New(block.Type + " not support")
}
