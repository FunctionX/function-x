package ios

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

var curve = elliptic.P256()

//GeneratePrivateKey generate private key by seed hex string
func GeneratePrivateKey(seed string) (string, error) {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return "", fmt.Errorf("invalid hex string")
	}
	priKey, err := ecdsa.GenerateKey(curve, bytes.NewReader(seedBytes))
	if err != nil {
		return "", fmt.Errorf("failed generating ECDSA key for [%v]: [%s]", curve, err)
	}
	return priKeyEncode(priKey), nil
}

func priKeyEncode(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
	return string(pemEncoded)
}

func priKeyDecode(pemEncoded string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemEncoded))
	key, err := dERToPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key1 := key.(*ecdsa.PrivateKey)
	return key1, nil
}

func dERToPrivateKey(der []byte) (key interface{}, err error) {
	if key, err = x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err = x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key.(type) {
		case *ecdsa.PrivateKey:
			return
		default:
			return nil, fmt.Errorf("found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}
	return nil, fmt.Errorf("invalid key type. The DER must contain an rsa.PrivateKey or ecdsa.PrivateKey")
}

//GetPublicKey get public key from private key
func GetPublicKey(privateKey string) (publicKey []byte, err error) {
	prikey, err := priKeyDecode(privateKey)
	if err != nil {
		return nil, err
	}
	return elliptic.Marshal(prikey.PublicKey.Curve, prikey.PublicKey.X, prikey.PublicKey.Y), nil
}

func hash(msg []byte) []byte {
	hash := sha256.New()
	hash.Write(msg)
	return hash.Sum(nil)
}

//Sign sign data with private key
func Sign(priKeyStr string, digest []byte) ([]byte, error) {
	priKey, err := priKeyDecode(priKeyStr)
	if err != nil {
		return nil, err
	}
	r, s, err := ecdsa.Sign(rand.Reader, priKey, hash(digest))
	if err != nil {
		return nil, err
	}
	s, _ = toLower(&priKey.PublicKey, s)
	bytes, err := marshalECDSASignature(r, s)
	return bytes, err
}

func toLower(k *ecdsa.PublicKey, s *big.Int) (*big.Int, bool) {
	if !isLower(s) {
		s.Sub(k.Params().N, s)
		return s, true
	}
	return s, false
}

func isLower(s *big.Int) bool {
	halfOrder := new(big.Int).Rsh(elliptic.P256().Params().N, 1)
	return s.Cmp(halfOrder) != 1
}

func marshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(eCDSASignature{r, s})
}

type eCDSASignature struct {
	R, S *big.Int
}

//Verify verify signature with public key
func Verify(pubB []byte, signature, digest []byte) error {
	x, y := elliptic.Unmarshal(curve, pubB)
	if x == nil {
		return errors.New("invalid " + curve.Params().Name + " public key")
	}
	pubKey := &ecdsa.PublicKey{Curve: curve, X: x, Y: y}

	r, s, err := unmarshalECDSASignature(signature)
	if err != nil {
		return fmt.Errorf("failed unmashalling signature [%s]", err)
	}
	lowS := isLower(s)
	if !lowS {
		return fmt.Errorf("invalid S. Must be smaller than half the order [%s][%s]", s, curve)
	}
	fmt.Println("r:", r)
	fmt.Println("s: ", s)
	if !ecdsa.Verify(pubKey, hash(digest), r, s) {
		return errors.New("verify failed")
	}
	return nil
}

func unmarshalECDSASignature(raw []byte) (*big.Int, *big.Int, error) {
	// Unmarshal
	sig := new(eCDSASignature)
	_, err := asn1.Unmarshal(raw, sig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed unmashalling signature [%s]", err)
	}

	// Validate sig
	if sig.R == nil {
		return nil, nil, errors.New("invalid signature, R must be different from nil")
	}
	if sig.S == nil {
		return nil, nil, errors.New("invalid signature, S must be different from nil")
	}

	if sig.R.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, R must be larger than zero")
	}
	if sig.S.Sign() != 1 {
		return nil, nil, errors.New("invalid signature, S must be larger than zero")
	}

	return sig.R, sig.S, nil
}

func GenCertificateRequest(priKey string) ([]byte, error) {
	priv, err := priKeyDecode(priKey)
	if err != nil {
		return nil, err
	}
	var tpl = x509.CertificateRequest{
		Subject:            pkix.Name{CommonName: "admin"},
		SignatureAlgorithm: x509.ECDSAWithSHA256,
		DNSNames:           []string{"localhost"},
	}
	csr, err := x509.CreateCertificateRequest(rand.Reader, &tpl, priv)
	if err != nil {
		return nil, err
	}
	block := pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csr,
	}
	csr = pem.EncodeToMemory(&block)
	return csr, nil
}
