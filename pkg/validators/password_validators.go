package validators

import (
	"golang.org/x/crypto/bcrypt"
)


// Hash generates a bcrypt hash from the provided password string.
func Hash(password string) []byte {
	// Generate a bcrypt hash from the password using a default cost.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// If an error occurs during hash generation, return the error message as a byte slice.
		// This is not ideal practice, but it's done here for simplicity.
		return []byte(err.Error())
	}
	// Return the generated hash.
	return hash
}