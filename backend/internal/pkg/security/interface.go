package security

type CipherInterface interface {
	Encrypt(pin, secret string) ([]byte, []byte, error)
	Decrypt(pin string, pinHash, secretHash []byte) (string, error)
}
