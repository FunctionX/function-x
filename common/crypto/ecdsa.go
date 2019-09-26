package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
)

var Curve = elliptic.P256()

func Hash(msg []byte) []byte {
	hash := sha256.New()
	hash.Write(msg)
	return hash.Sum(nil)
}

func KeyGen() (*ecdsa.PrivateKey, error) {
	priKey, err := ecdsa.GenerateKey(Curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed generating ECDSA key for [%v]: [%s]", Curve, err)
	}
	return priKey, nil
}

func KeyGenWithHash(hash io.Reader) (*ecdsa.PrivateKey, error) {
	priKey, err := ecdsa.GenerateKey(Curve, hash)
	if err != nil {
		return nil, fmt.Errorf("failed generating ECDSA key for [%v]: [%s]", Curve, err)
	}
	return priKey, nil
}

func SignPri(priKey *ecdsa.PrivateKey, digest []byte) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, priKey, digest)
	if err != nil {
		return nil, err
	}

	s, _ = toLower(&priKey.PublicKey, s)
	return MarshalECDSASignature(r, s)
}

func Sign(private string, info string, hash bool) ([]byte, error) {
	key, e := PriKeyDecode(private)
	if e != nil {
		return nil, e
	}
	content := []byte(info)
	if hash {
		content = Hash(content)
	}
	return SignPri(key, content)
}

func Verify(pubKey *ecdsa.PublicKey, signature, digest []byte) (bool, error) {
	r, s, err := UnmarshalECDSASignature(signature)
	if err != nil {
		return false, fmt.Errorf("failed unmashalling signature [%s]", err)
	}

	lowS := isLower(s)

	if !lowS {
		return false, fmt.Errorf("invalid S. Must be smaller than half the order [%s][%s]", s, Curve)
	}
	fmt.Println("r:", r)
	fmt.Println("s: ", s)
	return ecdsa.Verify(pubKey, digest, r, s), nil
}

func PriKeyEncode(privateKey *ecdsa.PrivateKey) string {
	x509Encoded, _ := x509.MarshalECPrivateKey(privateKey)
	pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: x509Encoded})
	return string(pemEncoded)
}

func PriKeyDecode(pemEncoded string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemEncoded))
	key, err := DERToPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	key1 := key.(*ecdsa.PrivateKey)
	return key1, nil
}

func DERToPrivateKey(der []byte) (key interface{}, err error) {
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

func PubKeyEncode(pub *ecdsa.PublicKey, encode ...string) string {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return ""
	}
	byte := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	if len(encode) > 0 {
		if encode[0] == "base64" {
			return base64.StdEncoding.EncodeToString(byte)
		}
	}
	return hex.EncodeToString(byte)
}

func PubKeyDecode(pub string, curve elliptic.Curve, decode ...string) (*ecdsa.PublicKey, error) {
	var bytes []byte
	var err error
	if len(decode) > 0 {
		if decode[0] == "base64" {
			bytes, err = base64.StdEncoding.DecodeString(pub)
		}

	} else {
		bytes, err = hex.DecodeString(pub)
	}
	fmt.Println(bytes)
	if err != nil {
		return nil, errors.New("invalid " + pub + "," + err.Error())
	}

	x, y := elliptic.Unmarshal(curve, bytes)
	if x == nil {
		return nil, errors.New("invalid " + curve.Params().Name + " public key")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

func toLower(k *ecdsa.PublicKey, s *big.Int) (*big.Int, bool) {
	if !isLower(s) {
		//s > n/2
		fmt.Println("before: ", s)
		s.Sub(k.Params().N, s)
		fmt.Println(" params: ", k.Params().N, s)
		return s, true
	}
	return s, false
}

//s > n/2 false
// s <= n/2 true
func isLower(s *big.Int) bool {
	halfOrder := new(big.Int).Rsh(elliptic.P256().Params().N, 1)
	fmt.Println("s vs n/2", s, halfOrder)
	return s.Cmp(halfOrder) != 1
}

func MarshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ECDSASignature{r, s})
}

func UnmarshalECDSASignature(raw []byte) (*big.Int, *big.Int, error) {
	// Unmarshal
	sig := new(ECDSASignature)
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

type ECDSASignature struct {
	R, S *big.Int
}

func Address(key ecdsa.PublicKey) string {
	raw := elliptic.Marshal(Curve, key.X, key.Y)
	// Hash it
	hash := sha256.New()
	hash.Write(raw)
	return hex.EncodeToString(hash.Sum(nil))
}
