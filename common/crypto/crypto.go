package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

const (
	// seed bytes
	SeedBytes = 64 // 512 bits
	// number of bits in a big.Word
	wordBits = 32 << (uint64(^big.Word(0)) >> 63)
	// number of bytes in a big.Word
	wordBytes = wordBits / 8
)

var secp256k1N, _ = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)

//GenerateKey generate private key
func GenerateKey(seed []byte) (privateKey *ecdsa.PrivateKey, err error) {
	if len(seed) != SeedBytes {
		return nil, fmt.Errorf("seed length must be 512 bits / 64 byte")
	}
	hmac512 := hmac.New(sha512.New, []byte("p2p seed"))
	hmac512.Write(seed)
	lr := hmac512.Sum(nil)
	secretKey := lr[:len(lr)/2]
	curve := S256()
	x, y := curve.ScalarBaseMult(secretKey)
	priv := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: new(big.Int).SetBytes(secretKey),
	}
	return priv, nil
}

// PublicKey to bytes
func PublicKey2Bytes(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(pub.Curve, pub.X, pub.Y)
}

//bytes to publickey
func Bytes2PublicKey(pub []byte, curve elliptic.Curve) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(curve, pub)
	if x == nil {
		return nil, errors.New("invalid " + curve.Params().Name + " public key")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

func String2PublicKey(key string) (*ecdsa.PublicKey, error) {
	return Bytes2PublicKey([]byte(key), S256())
}

//Secp256k1 PrivateKey to bytes
func PrivateKey2Bytes(priv *ecdsa.PrivateKey) []byte {
	if priv == nil {
		return nil
	}
	n := priv.Params().BitSize / 8

	if priv.D.BitLen()/8 >= n {
		return priv.D.Bytes()
	}
	ret := make([]byte, n)

	for _, d := range priv.D.Bits() {
		for j := 0; j < wordBytes && n > 0; j++ {
			n--
			ret[n] = byte(d)
			d >>= 8
		}
	}
	return ret
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

type Address [20]byte

// HexToAddress returns Address with byte values of s.
// If s is larger than len(h), s will be cropped from the left.
func HexToAddress(s string) Address { return BytesToAddress(FromHex(s)) }

// FromHex returns the bytes represented by the hexadecimal string s.
// s may be prefixed with "0x".
func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" || s[0:2] == "0X" {
			s = s[2:]
		}
	}
	if len(s)%2 == 1 {
		s = "0" + s
	}
	return Hex2Bytes(s)
}

// Hex2Bytes returns the bytes represented by the hexadecimal string str.
func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)
	return h
}

// Hex returns an EIP55-compliant hex string representation of the address.
func (a Address) Hex() string {
	unchecksummed := hex.EncodeToString(a[:])
	sha := sha3.NewLegacyKeccak256()
	sha.Write([]byte(unchecksummed))
	hash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

// SetBytes sets the address to the value of b.
// If b is larger than len(a) it will panic.
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-20:]
	}
	copy(a[20-len(b):], b)
}

// BytesToAddress returns Address with value b.
// If b is larger than len(h), b will be cropped from the left.
func BytesToAddress(b []byte) Address {
	var a Address
	a.SetBytes(b)
	return a
}

func PublicKeyToAddress(p *ecdsa.PublicKey) Address {
	pubBytes := FromECDSAPub(p)
	return BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

func PubKeyBytesToAddress(pb []byte) Address {
	return BytesToAddress(Keccak256(pb[1:])[12:])
}

//hex string to Secp256k1 PrivateKey
func HexStr2Secp256k1PrivateKey(hexkey string) (*ecdsa.PrivateKey, error) {
	d, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()

	if 8*len(d) != priv.Params().BitSize {
		return nil, fmt.Errorf("invalid length, need %d bits", priv.Params().BitSize)
	}
	priv.D = new(big.Int).SetBytes(d)

	//The priv.D must < N
	if priv.D.Cmp(secp256k1N) >= 0 {
		return nil, fmt.Errorf("invalid private key, >=N")
	}
	// The priv.D must not be zero or negative.
	if priv.D.Sign() <= 0 {
		return nil, fmt.Errorf("invalid private key, zero or negative")
	}

	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(d)
	if priv.PublicKey.X == nil {
		return nil, errors.New("invalid private key")
	}
	return priv, nil
}

func Sign(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	md5data := md5.New()
	md5data.Write(data)
	hash := md5data.Sum([]byte(""))

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, err
	}
	sign := append(r.Bytes(), s.Bytes()...)
	return sign, nil
}

func Verify(sign, data []byte, publicKey *ecdsa.PublicKey) bool {
	md5data := md5.New()
	md5data.Write(data)
	hash := md5data.Sum([]byte(""))
	if !ecdsa.Verify(publicKey, hash, new(big.Int).SetBytes(sign[:32]), new(big.Int).SetBytes(sign[32:])) {
		return false
	}
	return true
}
