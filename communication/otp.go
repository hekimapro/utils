package communication

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/hekimapro/utils/log"
)

func GenerateOTP() (int, error) {
	min := int64(100000) // Smallest 6-digit number
	max := int64(999999) // Largest 6-digit number

	// Generate a random number between 100000 and 999999
	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		log.Error(err.Error())
		return 0, fmt.Errorf("failed to generate OTP")
	}

	// Ensure it’s always 6 digits
	otp := int(n.Int64() + min)
	return otp, nil
}
