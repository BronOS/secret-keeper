package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"errors"
	"io"
)

type Aes struct {
	cipherKey string
}

func (a *Aes) getKey(pin string) []byte {
	return []byte(a.md5(a.cipherKey + pin))[:16]
}

func (a *Aes) md5(text string) []byte {
	h := md5.Sum([]byte(text))
	return h[:]
}

func (a *Aes) Encrypt(pin, secret string) ([]byte, []byte, error) {
	c, err := aes.NewCipher(a.getKey(pin))
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	return a.md5(pin), gcm.Seal(nonce, nonce, []byte(secret), nil), nil
}

func (a *Aes) Decrypt(pin string, pinHash, secretHash []byte) (string, error) {
	if !a.comparePin(pin, pinHash) {
		return "", errors.New("incorrect pin")
	}

	c, err := aes.NewCipher(a.getKey(pin))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(secretHash) < nonceSize {
		return "", errors.New("secret hash too short")
	}

	nonce, secretHash := secretHash[:nonceSize], secretHash[nonceSize:]
	secret, err := gcm.Open(nil, nonce, secretHash, nil)

	return string(secret), err
}

func (a *Aes) comparePin(pin string, pinHash []byte) bool {
	return bytes.Compare(a.md5(pin), pinHash) == 0
}

func NewAes(cipherKey string) *Aes {
	return &Aes{
		cipherKey: cipherKey,
	}
}
