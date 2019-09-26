package crypto

import (
	"crypto/rand"
	"fmt"
)

const (
	// NonceSize is the default NonceSize
	nonceSize = 24
)

// GetRandomBytes returns len random looking bytes
func getRandomBytes(len int) ([]byte, error) {
	key := make([]byte, len)

	// TODO: rand could fill less bytes then len
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("error getting random bytes: %s", err.Error())
	}

	return key, nil
}

// GetRandomNonce returns a random byte array of length NonceSize
func getRandomNonce() ([]byte, error) {
	return getRandomBytes(nonceSize)
}
