package encryption

import (
	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
	"golang.org/x/crypto/bcrypt"    // bcrypt provides password hashing and verification functions.
)

// CreateHash generates a bcrypt hash from a plain text password.
// Returns the hashed password as a string or an error if hashing fails.
func CreateHash(Password string) (string, error) {
	// Log the start of the password hashing process.
	log.Info("🔐 Generating bcrypt hash from password")

	// Generate a bcrypt hash using the default cost factor.
	HashedString, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		// Log and return an error if hashing fails.
		log.Error("❌ Failed to generate hash: " + err.Error())
		return "", err
	}

	// Log successful hash generation.
	log.Success("✅ Password hash created successfully")
	// Convert the hash to a string and return it.
	return string(HashedString), nil
}

// CompareWithHash verifies a plain text password against a bcrypt hash.
// Returns true if the password matches the hash, false otherwise.
func CompareWithHash(HashedString string, Password string) bool {
	// Log the start of the password verification process.
	log.Info("🔎 Verifying password against bcrypt hash")

	// Compare the provided password with the stored hash.
	err := bcrypt.CompareHashAndPassword([]byte(HashedString), []byte(Password))
	if err != nil {
		// Log and return false if the password does not match.
		log.Error("❌ Password does not match hash")
		return false
	}

	// Log successful password verification.
	log.Success("✅ Password verification successful")
	return true
}