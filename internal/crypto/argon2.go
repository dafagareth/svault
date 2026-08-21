package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	SaltSize = 16
	KeySize  = 32
)

// exported vars (not const) so tests can lower them for speed without forking the package
var (
	ArgonTime    uint32 = 3
	ArgonMemory  uint32 = 64 * 1024
	ArgonThreads uint8  = 4
)

func NewSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate salt: %w", err)
	}
	return salt, nil
}

func DeriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, ArgonTime, ArgonMemory, ArgonThreads, KeySize)
}
