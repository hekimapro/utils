package utils

import "golang.org/x/crypto/bcrypt"

// CreateHash generates a bcrypt hash from the provided password.
func CreateHash(Password string) (string, error) {
	// GenerateFromPassword returns the bcrypt hash of the password
	HashedString, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(HashedString), nil
}

// CompareWithHash compares a password with a bcrypt hash.
// It returns true if the password matches the hash, and false otherwise.
func CompareWithHash(HashedString string, Password string) bool {
	// CompareHashAndPassword compares the bcrypt hashed password with the entered password
	err := bcrypt.CompareHashAndPassword([]byte(HashedString), []byte(Password))
	return err == nil
}
