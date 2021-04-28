package security

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAes_Decrypt(t *testing.T) {
	cipher := NewAes("lia?hG1d-p9\\82U3=y4!")
	pin := "123"
	secret := "&jU~q_23u01<"

	pinHash, secretHash, eerr := cipher.Encrypt(pin, secret)

	assert.NoError(t, eerr)

	dsecret, derr := cipher.Decrypt(pin, pinHash, secretHash)

	assert.NoError(t, derr)
	assert.Equal(t, secret, dsecret)
}
