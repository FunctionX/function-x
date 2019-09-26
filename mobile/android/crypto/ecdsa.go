package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
)

var curve = elliptic.P256()

func hash(msg []byte) []byte {
	hash := sha256.New()
	hash.Write(msg)
	return hash.Sum(nil)
}

func KeyGen() (string, error) {
	priKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed generating ECDSA key for [%v]: [%s]", curve, err)
	}
	return priKeyEncode(priKey), nil
}

func KeyGenByMnemonic(mnemonic []byte) (string, error) {
	priKey, err := ecdsa.GenerateKey(curve, bytes.NewReader(mnemonic))
	if err != nil {
		return "", fmt.Errorf("failed generating ECDSA key for [%v]: [%s]", curve, err)
	}
	return priKeyEncode(priKey), nil
}

func PublicKey(privateKey string) (string, error) {
	key, e := priKeyDecode(privateKey)
	if e != nil {
		return "", e
	}
	pub := &key.PublicKey
	if pub == nil || pub.X == nil || pub.Y == nil {
		return "", fmt.Errorf("failed to get public key cause of empty private key")
	}
	byte := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	return hex.EncodeToString(byte), nil
}

func signPri(priKey *ecdsa.PrivateKey, digest []byte) (string, error) {
	r, s, err := ecdsa.Sign(rand.Reader, priKey, digest)
	if err != nil {
		return "", err
	}

	s, _ = toLower(&priKey.PublicKey, s)
	bytes, err := marshalECDSASignature(r, s)
	return string(bytes), err
}

func Sign(private string, info string, hash1 bool) (string, error) {
	key, e := priKeyDecode(private)
	if e != nil {
		return "", e
	}
	content := []byte(info)
	if hash1 {
		content = hash(content)
	}
	return signPri(key, content)
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
			return nil, errors.New("found unknown private key type in PKCS#8 wrapping")
		}
	}

	if key, err = x509.ParseECPrivateKey(der); err == nil {
		return
	}
	return nil, errors.New("invalid key type. The DER must contain an rsa.PrivateKey or ecdsa.PrivateKey")
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
	return asn1.Marshal(ECDSASignature{r, s})
}

type ECDSASignature struct {
	R, S *big.Int
}
