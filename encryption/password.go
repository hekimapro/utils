package encryption

import "golang.org/x/crypto/bcrypt"

// CreateHash generates a bcrypt hash from a plain text password
// Uses the default bcrypt cost factor to create a secure hash
// Returns the hashed password as a string or an error if hashing fails
func CreateHash(Password string) (string, error) {
	// Generate a bcrypt hash from the password with the default cost
	HashedString, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		// Return an empty string and the error if hashing fails
		return "", err
	}
	// Convert the hashed bytes to a string and return
	return string(HashedString), nil
}

// CompareWithHash verifies a plain text password against a bcrypt hash
// Checks if the provided password matches the hashed password
// Returns true if the password matches, false otherwise
func CompareWithHash(HashedString string, Password string) bool {
	// Compare the hashed password with the plain text password
	err := bcrypt.CompareHashAndPassword([]byte(HashedString), []byte(Password))
	// Return true if the passwords match (no error), false if they don’t
	return err == nil
}