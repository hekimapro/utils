package communication

import (
	"crypto/rand" // rand provides cryptographically secure random number generation.
	"fmt"         // fmt provides formatting and printing functions.
	"math/big"    // big provides support for large integer arithmetic.

	"github.com/hekimapro/utils/log" // log provides colored logging utilities.
)

// GenerateOTP generates a secure 6-digit One-Time Password (OTP).
// Returns the OTP and an error (if any occurs during generation).
func GenerateOTP() (int, error) {
	const (
		min = int64(100000) // Smallest 6-digit number for OTP range.
		max = int64(999999) // Largest 6-digit number for OTP range.
	)

	// Log the start of the OTP generation process.
	log.Info("🔐 Generating a secure 6-digit OTP")

	// Generate a cryptographically secure random number in the range [0, max-min].
	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		// Log and return an error if random number generation fails.
		log.Error(fmt.Sprintf("❌ Failed to generate secure random number: %v", err))
		return 0, fmt.Errorf("failed to generate OTP")
	}

	// Shift the random number to the 6-digit range [100000, 999999].
	otp := int(n.Int64() + min)

	// Log successful OTP generation with the generated value.
	log.Success(fmt.Sprintf("✅ OTP generated successfully: %d", otp))
	return otp, nil
}
