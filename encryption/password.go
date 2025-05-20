package encryption

import (
	"github.com/hekimapro/utils/log"
	"golang.org/x/crypto/bcrypt"
)

// CreateHash generates a bcrypt hash from a plain text password
func CreateHash(Password string) (string, error) {
	log.Info("Generating bcrypt hash")

	HashedString, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("Failed to generate hash: " + err.Error())
		return "", err
	}

	log.Success("Password hash created successfully")
	return string(HashedString), nil
}

// CompareWithHash verifies a plain text password against a bcrypt hash
func CompareWithHash(HashedString string, Password string) bool {
	log.Info("Comparing password with hash")

	err := bcrypt.CompareHashAndPassword([]byte(HashedString), []byte(Password))
	if err != nil {
		log.Error("Password comparison failed: " + err.Error())
		return false
	}

	log.Success("Password matches hash")
	return true
}
