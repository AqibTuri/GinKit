// Package password hashes and verifies passwords with bcrypt (one-way; not reversible encryption).
package password

import (
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// Hash returns a bcrypt hash of the plaintext password (one-way; not encryption).
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify checks plaintext against a bcrypt hash.
func Verify(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
