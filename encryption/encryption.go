package encryption

import (
	"bytes"           // bytes provides utilities for byte slice manipulation (e.g., padding).
	"context"         // context provides support for cancellation and timeouts.
	"crypto/aes"      // aes provides AES encryption and decryption functionality.
	"crypto/cipher"   // cipher provides block cipher modes like CBC.
	"crypto/rand"     // rand provides cryptographically secure random number generation.
	"encoding/base64" // base64 provides Base64 encoding/decoding.
	"encoding/hex"    // hex provides hexadecimal encoding/decoding.
	"encoding/json"   // json provides JSON encoding/decoding.
	"errors"          // errors provides error creation utilities.
	"fmt"
	"io"
	"strings"
	"time" // time provides functionality for timeouts and durations.

	"github.com/hekimapro/utils/helpers"
	"github.com/hekimapro/utils/log"    // log provides colored logging utilities.
	"github.com/hekimapro/utils/models" // models contains data structures for encryption payloads.
)

// pad applies PKCS7 padding to the plaintext to align with AES block size.
// Returns the padded byte slice.
func pad(src []byte, blockSize int) []byte {
	// Calculate padding size to make the input length a multiple of blockSize.
	padding := blockSize - len(src)%blockSize
	// Create padding bytes with the value equal to the padding length.
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	// Append padding to the source data.
	return append(src, padText...)
}

// unpad removes PKCS7 padding from the decrypted plaintext.
// Returns the unpadded data or an error if the padding is invalid.
func unpad(src []byte) ([]byte, error) {
	// Check if input is empty to prevent invalid access.
	length := len(src)
	if length == 0 {
		return nil, errors.New("invalid padding size: empty input")
	}
	// Extract padding length from the last byte.
	padding := int(src[length-1])
	// Validate padding length to ensure it's within bounds.
	if padding > length || padding == 0 {
		return nil, errors.New("invalid padding size")
	}
	// Return the data with padding removed.
	return src[:length-padding], nil
}

// getEncryptionConfig retrieves encryption configuration with context support.
func getEncryptionConfig(ctx context.Context) (*models.EncryptionConfig, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "encryption config loading cancelled")
	default:
		// Continue with config loading
	}

	var missing []string
	config := &models.EncryptionConfig{
		EncryptionType:       helpers.GetENVValue("encryption type"),
		EncryptionKey:        helpers.GetENVValue("encryption key"),
		InitializationVector: helpers.GetENVValue("initialization vector"),
	}

	if config.EncryptionKey == "" {
		missing = append(missing, "ENCRYPTION_KEY")
	}

	if config.EncryptionType == "" {
		missing = append(missing, "ENCRYPTION_TYPE")
	}

	if config.InitializationVector == "" {
		missing = append(missing, "INITIALIZATION_VECTOR")
	}

	if len(missing) > 0 {
		return config, fmt.Errorf(".env file is missing required encryption config(s): %s", strings.Join(missing, ", "))
	}

	return config, nil
}

// validateEncryptionConfig validates encryption configuration parameters.
func validateEncryptionConfig(config *models.EncryptionConfig) error {
	// Validate that the initialization vector is exactly 16 bytes (AES block size).
	if len(config.InitializationVector) != aes.BlockSize {
		return errors.New("initialization vector must be exactly 16 bytes long")
	}

	// Validate that the encryption type is either "base64" or "hex".
	if config.EncryptionType != "base64" && config.EncryptionType != "hex" {
		return errors.New("invalid encryption type (use 'base64' or 'hex')")
	}

	// Validate encryption key length (should be 16, 24, or 32 bytes for AES)
	keyLength := len(config.EncryptionKey)
	if keyLength != 16 && keyLength != 24 && keyLength != 32 {
		return fmt.Errorf("encryption key must be 16, 24, or 32 bytes long, got %d bytes", keyLength)
	}

	return nil
}

// generateRandomIV generates a cryptographically secure random initialization vector.
func generateRandomIV() ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, helpers.WrapError(err, "failed to generate random IV")
	}
	return iv, nil
}

// Encrypt encrypts data using AES in CBC mode and returns an encoded payload.
// Supports Base64 or hex encoding for the ciphertext.
// Returns the encrypted payload or an error if encryption fails.
func Encrypt(data interface{}) (*models.EncryptReturnType, error) {
	// Create context with timeout for encryption operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return encryptWithContext(ctx, data)
}

// encryptWithContext is the internal implementation with context support.
func encryptWithContext(ctx context.Context, data interface{}) (*models.EncryptReturnType, error) {
	// Log the start of the encryption process.
	log.Info("🔐 Starting encryption process")

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "encryption cancelled before start")
	default:
		// Continue with encryption
	}

	config, err := getEncryptionConfig(ctx)
	if err != nil {
		log.Error("❌ " + err.Error())
		return nil, err
	}

	// Validate configuration
	if err := validateEncryptionConfig(config); err != nil {
		log.Error("❌ " + err.Error())
		return nil, err
	}

	// Check context cancellation after config validation
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "encryption cancelled after config validation")
	default:
		// Continue with encryption
	}

	// Marshal the input data to JSON for encryption.
	dataToEncrypt, err := json.Marshal(data)
	if err != nil {
		log.Error("❌ Failed to marshal input data: " + err.Error())
		return nil, helpers.WrapError(err, "failed to marshal input data")
	}

	// Check context cancellation after marshaling
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "encryption cancelled after data marshaling")
	default:
		// Continue with encryption
	}

	// Initialize AES cipher with the provided key.
	block, err := aes.NewCipher([]byte(config.EncryptionKey))
	if err != nil {
		log.Error("❌ Failed to initialize AES cipher: " + err.Error())
		return nil, helpers.WrapError(err, "failed to initialize AES cipher")
	}

	// Apply PKCS7 padding to the data to match AES block size.
	log.Info("📦 Padding data")
	paddedData := pad(dataToEncrypt, aes.BlockSize)

	// Check context cancellation after padding
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "encryption cancelled after padding")
	default:
		// Continue with encryption
	}

	// Perform AES-CBC encryption.
	log.Info("🔁 Performing AES-CBC encryption")
	mode := cipher.NewCBCEncrypter(block, []byte(config.InitializationVector))
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	// Check context cancellation after encryption
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "encryption cancelled after encryption")
	default:
		// Continue with encoding
	}

	// Encode the ciphertext based on the specified encoding type.
	var encryptedPayload string
	if config.EncryptionType == "base64" {
		encryptedPayload = base64.StdEncoding.EncodeToString(ciphertext)
	} else {
		encryptedPayload = hex.EncodeToString(ciphertext)
	}

	// Log successful encryption.
	log.Success("✅ Data encrypted successfully")
	return &models.EncryptReturnType{Payload: encryptedPayload}, nil
}

// Decrypt decrypts AES-encrypted data in CBC mode and returns the original data.
// Supports Base64 or hex-encoded input.
// Returns the decrypted data or an error if decryption fails.
func Decrypt(encryptedData models.EncryptReturnType) (interface{}, error) {
	// Create context with timeout for decryption operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return decryptWithContext(ctx, encryptedData)
}

// decryptWithContext is the internal implementation with context support.
func decryptWithContext(ctx context.Context, encryptedData models.EncryptReturnType) (interface{}, error) {
	// Log the start of the decryption process.
	log.Info("🔓 Starting decryption process")

	// Check context cancellation before starting
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "decryption cancelled before start")
	default:
		// Continue with decryption
	}

	config, err := getEncryptionConfig(ctx)
	if err != nil {
		log.Error("❌ " + err.Error())
		return nil, err
	}

	// Validate configuration
	if err := validateEncryptionConfig(config); err != nil {
		log.Error("❌ " + err.Error())
		return nil, err
	}

	// Check context cancellation after config validation
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "decryption cancelled after config validation")
	default:
		// Continue with decryption
	}

	// Decode the encrypted payload based on the specified encoding type.
	log.Info("📥 Decoding encrypted payload")
	var ciphertext []byte

	if config.EncryptionType == "base64" {
		ciphertext, err = base64.StdEncoding.DecodeString(encryptedData.Payload)
	} else {
		ciphertext, err = hex.DecodeString(encryptedData.Payload)
	}
	if err != nil {
		log.Error("❌ Failed to decode payload: " + err.Error())
		return nil, helpers.WrapError(err, "failed to decode payload")
	}

	// Check context cancellation after decoding
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "decryption cancelled after payload decoding")
	default:
		// Continue with decryption
	}

	// Initialize AES cipher with the provided key.
	block, err := aes.NewCipher([]byte(config.EncryptionKey))
	if err != nil {
		log.Error("❌ Failed to initialize AES cipher: " + err.Error())
		return nil, helpers.WrapError(err, "failed to initialize AES cipher")
	}

	// Perform AES-CBC decryption.
	log.Info("🔁 Performing AES-CBC decryption")
	mode := cipher.NewCBCDecrypter(block, []byte(config.InitializationVector))
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Check context cancellation after decryption
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "decryption cancelled after decryption")
	default:
		// Continue with padding removal
	}

	// Remove PKCS7 padding from the decrypted data.
	log.Info("🧹 Removing padding")
	plaintext, err = unpad(plaintext)
	if err != nil {
		log.Error("❌ Padding removal failed: " + err.Error())
		return nil, helpers.WrapError(err, "padding removal failed")
	}

	// Check context cancellation after padding removal
	select {
	case <-ctx.Done():
		return nil, helpers.WrapError(ctx.Err(), "decryption cancelled after padding removal")
	default:
		// Continue with unmarshaling
	}

	// Unmarshal the decrypted JSON data into an interface.
	log.Info("🧩 Unmarshaling decrypted data")
	var decryptedData interface{}
	err = json.Unmarshal(plaintext, &decryptedData)
	if err != nil {
		log.Error("❌ JSON unmarshaling failed: " + err.Error())
		return nil, helpers.WrapError(err, "JSON unmarshaling failed")
	}

	// Log successful decryption.
	log.Success("✅ Data decrypted successfully")
	return decryptedData, nil
}

// EncryptString is a convenience function for encrypting string data.
func EncryptString(data string) (*models.EncryptReturnType, error) {
	return Encrypt(data)
}

// DecryptString is a convenience function for decrypting to string data.
func DecryptString(encryptedData models.EncryptReturnType) (string, error) {
	result, err := Decrypt(encryptedData)
	if err != nil {
		return "", err
	}

	str, ok := result.(string)
	if !ok {
		return "", helpers.CreateError("decrypted data is not a string")
	}

	return str, nil
}

// EncryptBytes is a convenience function for encrypting byte data.
func EncryptBytes(data []byte) (*models.EncryptReturnType, error) {
	return Encrypt(data)
}

// DecryptBytes is a convenience function for decrypting to byte data.
func DecryptBytes(encryptedData models.EncryptReturnType) ([]byte, error) {
	result, err := Decrypt(encryptedData)
	if err != nil {
		return nil, err
	}

	bytes, ok := result.([]byte)
	if !ok {
		// Try to convert if it's a slice of interfaces
		if slice, ok := result.([]interface{}); ok {
			bytes := make([]byte, len(slice))
			for i, v := range slice {
				if b, ok := v.(float64); ok {
					bytes[i] = byte(b)
				} else {
					return nil, helpers.CreateError("decrypted data cannot be converted to bytes")
				}
			}
			return bytes, nil
		}
		return nil, helpers.CreateError("decrypted data is not bytes")
	}

	return bytes, nil
}

// GenerateEncryptionKey generates a cryptographically secure random encryption key.
func GenerateEncryptionKey(keySize int) (string, error) {
	if keySize != 16 && keySize != 24 && keySize != 32 {
		return "", helpers.CreateError("key size must be 16, 24, or 32 bytes")
	}

	key := make([]byte, keySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", helpers.WrapError(err, "failed to generate encryption key")
	}

	return string(key), nil
}

// GenerateIV generates a cryptographically secure random initialization vector.
func GenerateIV() (string, error) {
	iv, err := generateRandomIV()
	if err != nil {
		return "", err
	}
	return string(iv), nil
}

// ValidateEncryptionKey validates if a key is suitable for AES encryption.
func ValidateEncryptionKey(key string) error {
	length := len(key)
	if length != 16 && length != 24 && length != 32 {
		return helpers.CreateErrorf("encryption key must be 16, 24, or 32 bytes long, got %d bytes", length)
	}
	return nil
}

// ValidateIV validates if an initialization vector is suitable for AES encryption.
func ValidateIV(iv string) error {
	if len(iv) != aes.BlockSize {
		return helpers.CreateErrorf("initialization vector must be exactly %d bytes long", aes.BlockSize)
	}
	return nil
}

// GetSupportedKeySizes returns the supported AES key sizes.
func GetSupportedKeySizes() []int {
	return []int{16, 24, 32}
}

// GetSupportedEncryptionTypes returns the supported encryption encoding types.
func GetSupportedEncryptionTypes() []string {
	return []string{"base64", "hex"}
}
