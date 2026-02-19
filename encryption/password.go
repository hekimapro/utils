package encryption

import (
	"context" // context provides support for cancellation and timeouts.
	"fmt"     // fmt provides formatting and printing functions.
	"time"    // time provides functionality for timeouts and durations.

	"github.com/hekimapro/utils/helpers" // helpers provides utility functions.
	"github.com/hekimapro/utils/log"     // log provides colored logging utilities.
	"golang.org/x/crypto/bcrypt"         // bcrypt provides password hashing and verification functions.
)

// CreateHash generates a bcrypt hash from a plain text password.
// Returns the hashed password as a string or an error if hashing fails.
func CreateHash(Password string) (string, error) {
	// Create context with timeout for hashing operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return createHashWithContext(ctx, Password)
}

// createHashWithContext is the internal implementation with context support.
func createHashWithContext(ctx context.Context, Password string) (string, error) {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return "", helpers.WrapError(ctx.Err(), "password hashing cancelled before start")
	default:
		// Continue with hashing
	}

	// Log the start of the password hashing process.
	log.Info("🔐 Generating bcrypt hash from password")

	// Validate input
	if Password == "" {
		log.Error("❌ Cannot hash empty password")
		return "", helpers.CreateError("password cannot be empty")
	}

	// Use a channel to handle the bcrypt operation with context
	resultChan := make(chan hashResult, 1)

	go func() {
		// Generate a bcrypt hash using the default cost factor.
		HashedString, err := bcrypt.GenerateFromPassword([]byte(Password), bcrypt.DefaultCost)
		resultChan <- hashResult{hash: string(HashedString), err: err}
	}()

	// Wait for either the result or context cancellation
	select {
	case <-ctx.Done():
		// Context was cancelled or timed out
		log.Warning("⚠️ Password hashing operation cancelled or timed out")
		return "", helpers.WrapError(ctx.Err(), "password hashing cancelled")
	case result := <-resultChan:
		if result.err != nil {
			// Log and return an error if hashing fails.
			log.Error("❌ Failed to generate hash: " + result.err.Error())
			return "", helpers.WrapError(result.err, "failed to generate password hash")
		}

		// Log successful hash generation.
		log.Success("✅ Password hash created successfully")
		// Convert the hash to a string and return it.
		return result.hash, nil
	}
}

// CompareWithHash verifies a plain text password against a bcrypt hash.
// Returns true if the password matches the hash, false otherwise.
func CompareWithHash(HashedString string, Password string) bool {
	// Create context with timeout for verification operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return compareWithHashContext(ctx, HashedString, Password)
}

// compareWithHashContext is the internal implementation with context support.
func compareWithHashContext(ctx context.Context, HashedString string, Password string) bool {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		log.Warning("⚠️ Password verification cancelled before start")
		return false
	default:
		// Continue with verification
	}

	// Log the start of the password verification process.
	log.Info("🔎 Verifying password against bcrypt hash")

	// Validate inputs
	if HashedString == "" {
		log.Error("❌ Cannot verify with empty hash")
		return false
	}
	if Password == "" {
		log.Error("❌ Cannot verify empty password")
		return false
	}

	// Use a channel to handle the bcrypt comparison with context
	resultChan := make(chan bool, 1)

	go func() {
		// Compare the provided password with the stored hash.
		err := bcrypt.CompareHashAndPassword([]byte(HashedString), []byte(Password))
		resultChan <- (err == nil)
	}()

	// Wait for either the result or context cancellation
	select {
	case <-ctx.Done():
		// Context was cancelled or timed out
		log.Warning("⚠️ Password verification cancelled or timed out")
		return false
	case result := <-resultChan:
		if !result {
			// Log and return false if the password does not match.
			log.Error("❌ Password does not match hash")
			return false
		}

		// Log successful password verification.
		log.Success("✅ Password verification successful")
		return true
	}
}

// hashResult is a helper struct to pass hash results through channels.
type hashResult struct {
	hash string
	err  error
}

// CreateHashWithCost generates a bcrypt hash with a custom cost factor.
// Higher cost factors are more secure but slower.
func CreateHashWithCost(Password string, cost int) (string, error) {
	// Create context with timeout for hashing operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return createHashWithCostContext(ctx, Password, cost)
}

// createHashWithCostContext is the internal implementation with context support and custom cost.
func createHashWithCostContext(ctx context.Context, Password string, cost int) (string, error) {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return "", helpers.WrapError(ctx.Err(), "password hashing with custom cost cancelled before start")
	default:
		// Continue with hashing
	}

	// Validate cost factor
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		log.Warning(fmt.Sprintf("⚠️ Cost factor %d is outside recommended range (%d-%d), using default",
			cost, bcrypt.MinCost, bcrypt.MaxCost))
		cost = bcrypt.DefaultCost
	}

	log.Info(fmt.Sprintf("🔐 Generating bcrypt hash with cost factor %d", cost))

	// Validate input
	if Password == "" {
		log.Error("❌ Cannot hash empty password")
		return "", helpers.CreateError("password cannot be empty")
	}

	// Use a channel to handle the bcrypt operation with context
	resultChan := make(chan hashResult, 1)

	go func() {
		// Generate a bcrypt hash using the specified cost factor.
		HashedString, err := bcrypt.GenerateFromPassword([]byte(Password), cost)
		resultChan <- hashResult{hash: string(HashedString), err: err}
	}()

	// Wait for either the result or context cancellation
	select {
	case <-ctx.Done():
		// Context was cancelled or timed out
		log.Warning("⚠️ Password hashing with custom cost cancelled or timed out")
		return "", helpers.WrapError(ctx.Err(), "password hashing with custom cost cancelled")
	case result := <-resultChan:
		if result.err != nil {
			// Log and return an error if hashing fails.
			log.Error("❌ Failed to generate hash with custom cost: " + result.err.Error())
			return "", helpers.WrapError(result.err, "failed to generate password hash with custom cost")
		}

		// Log successful hash generation.
		log.Success("✅ Password hash created successfully with custom cost")
		// Convert the hash to a string and return it.
		return result.hash, nil
	}
}

// IsHashValid checks if a string appears to be a valid bcrypt hash.
func IsHashValid(hashedString string) bool {
	if hashedString == "" {
		return false
	}

	// Basic validation: bcrypt hashes start with $2a$, $2b$, $2x$, or $2y$
	if len(hashedString) < 60 {
		return false
	}

	prefix := hashedString[:4]
	return prefix == "$2a$" || prefix == "$2b$" || prefix == "$2x$" || prefix == "$2y$"
}

// GetHashInfo returns basic information about a bcrypt hash.
func GetHashInfo(hashedString string) (cost int, err error) {
	if !IsHashValid(hashedString) {
		return 0, helpers.CreateError("invalid bcrypt hash format")
	}

	// Extract cost from hash (format: $2a$cost$...)
	if len(hashedString) < 7 {
		return 0, helpers.CreateError("invalid bcrypt hash format")
	}

	// Cost is between the 3rd and 5th characters after the prefix
	costStr := hashedString[4:6]
	_, err = fmt.Sscanf(costStr, "%d", &cost)
	if err != nil {
		return 0, helpers.WrapError(err, "failed to parse cost from bcrypt hash")
	}

	return cost, nil
}

// NeedsRehash checks if a hash needs to be rehashed with a higher cost factor.
func NeedsRehash(hashedString string, minCost int) (bool, error) {
	if !IsHashValid(hashedString) {
		return false, helpers.CreateError("invalid bcrypt hash format")
	}

	currentCost, err := GetHashInfo(hashedString)
	if err != nil {
		return false, err
	}

	// Validate minCost parameter
	if minCost < bcrypt.MinCost || minCost > bcrypt.MaxCost {
		return false, helpers.CreateErrorf("minCost must be between %d and %d", bcrypt.MinCost, bcrypt.MaxCost)
	}

	return currentCost < minCost, nil
}

// GetRecommendedCost returns the recommended bcrypt cost factor for the current system.
// This can be used to automatically adjust cost based on system performance.
func GetRecommendedCost() int {
	// Start with default cost
	cost := bcrypt.DefaultCost

	// In a real implementation, you might want to benchmark the system
	// and adjust the cost factor accordingly. For now, we return the default.

	return cost
}

// HashAndVerify generates a hash and immediately verifies it against the original password.
// This is useful for ensuring the hash was generated correctly.
func HashAndVerify(Password string) (string, error) {
	// Create context with timeout for combined operation
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return hashAndVerifyContext(ctx, Password)
}

// hashAndVerifyContext is the internal implementation with context support.
func hashAndVerifyContext(ctx context.Context, Password string) (string, error) {
	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return "", helpers.WrapError(ctx.Err(), "hash and verify operation cancelled before start")
	default:
		// Continue with operation
	}

	log.Info("🔐 Generating and verifying password hash")

	// Generate the hash
	hashed, err := createHashWithContext(ctx, Password)
	if err != nil {
		return "", helpers.WrapError(err, "failed to generate hash for verification")
	}

	// Check context cancellation after hashing
	select {
	case <-ctx.Done():
		return "", helpers.WrapError(ctx.Err(), "hash and verify operation cancelled after hashing")
	default:
		// Continue with verification
	}

	// Verify the hash
	if !compareWithHashContext(ctx, hashed, Password) {
		return "", helpers.CreateError("generated hash failed verification against original password")
	}

	log.Success("✅ Password hash generated and verified successfully")
	return hashed, nil
}
