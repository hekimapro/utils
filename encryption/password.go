package encryption

import "golang.org/x/crypto/bcrypt"

// CreateHash generates a bcrypt hash from the provided password.
// Password: The plain text password to be hashed.
// Returns the hashed password string, or an error if the hashing fails.
func CreateHash(Password string) (string, error) {
	// Generate a bcrypt hash of the password using the default cost factor.
	HashedString, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		// Return an empty string and the error if hashing fails.
		return "", err
	}
	// Return the hashed password as a string.
	return string(HashedString), nil
}

// CompareWithHash compares a password with a bcrypt hash.
// HashedString: The bcrypt hashed password.
// Password: The plain text password to be checked.
// Returns true if the password matches the hash, false otherwise.
func CompareWithHash(HashedString string, Password string) bool {
	// Compare the hashed password with the plain text password.
	err := bcrypt.CompareHashAndPassword([]byte(HashedString), []byte(Password))
	// Return true if no error (passwords match), false if there is an error.
	return err == nil
}
