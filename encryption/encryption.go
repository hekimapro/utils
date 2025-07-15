package encryption

import (
	"bytes"           // bytes provides utilities for byte slice manipulation (e.g., padding).
	"crypto/aes"      // aes provides AES encryption and decryption functionality.
	"crypto/cipher"   // cipher provides block cipher modes like CBC.
	"encoding/base64" // base64 provides Base64 encoding/decoding.
	"encoding/hex"    // hex provides hexadecimal encoding/decoding.
	"encoding/json"   // json provides JSON encoding/decoding.
	"errors"          // errors provides error creation utilities.

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

// Encrypt encrypts data using AES in CBC mode and returns an encoded payload.
// Supports Base64 or hex encoding for the ciphertext.
// Returns the encrypted payload or an error if encryption fails.
func Encrypt(data interface{}, encryptionKey, initializationVector, encryptionType string) (models.EncryptReturnType, error) {
	// Log the start of the encryption process.
	log.Info("🔐 Starting encryption process")

	// Validate that the encryption key is non-empty.
	if encryptionKey == "" {
		log.Error("❌ Encryption key is empty")
		return models.EncryptReturnType{}, errors.New("encryption key must be a non-empty string")
	}

	// Validate that the initialization vector is exactly 16 bytes (AES block size).
	if len(initializationVector) != aes.BlockSize {
		log.Error("❌ Initialization vector must be 16 bytes")
		return models.EncryptReturnType{}, errors.New("initialization vector must be exactly 16 bytes long")
	}

	// Validate that the encryption type is either "base64" or "hex".
	if encryptionType != "base64" && encryptionType != "hex" {
		log.Error("❌ Unsupported encryption type")
		return models.EncryptReturnType{}, errors.New("invalid encryption type (use 'base64' or 'hex')")
	}

	// Marshal the input data to JSON for encryption.
	dataToEncrypt, err := json.Marshal(data)
	if err != nil {
		log.Error("❌ Failed to marshal input data: " + err.Error())
		return models.EncryptReturnType{}, err
	}

	// Initialize AES cipher with the provided key.
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		log.Error("❌ Failed to initialize AES cipher: " + err.Error())
		return models.EncryptReturnType{}, err
	}

	// Apply PKCS7 padding to the data to match AES block size.
	log.Info("📦 Padding data")
	paddedData := pad(dataToEncrypt, aes.BlockSize)

	// Perform AES-CBC encryption.
	log.Info("🔁 Performing AES-CBC encryption")
	mode := cipher.NewCBCEncrypter(block, []byte(initializationVector))
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	// Encode the ciphertext based on the specified encoding type.
	var encryptedPayload string
	if encryptionType == "base64" {
		encryptedPayload = base64.StdEncoding.EncodeToString(ciphertext)
	} else {
		encryptedPayload = hex.EncodeToString(ciphertext)
	}

	// Log successful encryption.
	log.Success("✅ Data encrypted successfully")
	return models.EncryptReturnType{Payload: encryptedPayload}, nil
}

// Decrypt decrypts AES-encrypted data in CBC mode and returns the original data.
// Supports Base64 or hex-encoded input.
// Returns the decrypted data or an error if decryption fails.
func Decrypt(encryptedData models.EncryptReturnType, encryptionKey, initializationVector, encryptionType string) (interface{}, error) {
	// Log the start of the decryption process.
	log.Info("🔓 Starting decryption process")

	// Validate that the encryption key is non-empty.
	if encryptionKey == "" {
		log.Error("❌ Encryption key is empty")
		return nil, errors.New("encryption key must be a non-empty string")
	}

	// Validate that the initialization vector is exactly 16 bytes (AES block size).
	if len(initializationVector) != aes.BlockSize {
		log.Error("❌ Initialization vector must be 16 bytes")
		return nil, errors.New("initialization vector must be exactly 16 bytes long")
	}

	// Validate that the encryption type is either "base64" or "hex".
	if encryptionType != "base64" && encryptionType != "hex" {
		log.Error("❌ Unsupported encryption type")
		return nil, errors.New("invalid encryption type")
	}

	// Decode the encrypted payload based on the specified encoding type.
	log.Info("📥 Decoding encrypted payload")
	var ciphertext []byte
	var err error
	if encryptionType == "base64" {
		ciphertext, err = base64.StdEncoding.DecodeString(encryptedData.Payload)
	} else {
		ciphertext, err = hex.DecodeString(encryptedData.Payload)
	}
	if err != nil {
		log.Error("❌ Failed to decode payload: " + err.Error())
		return nil, err
	}

	// Initialize AES cipher with the provided key.
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		log.Error("❌ Failed to initialize AES cipher: " + err.Error())
		return nil, err
	}

	// Perform AES-CBC decryption.
	log.Info("🔁 Performing AES-CBC decryption")
	mode := cipher.NewCBCDecrypter(block, []byte(initializationVector))
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding from the decrypted data.
	log.Info("🧹 Removing padding")
	plaintext, err = unpad(plaintext)
	if err != nil {
		log.Error("❌ Padding removal failed: " + err.Error())
		return nil, err
	}

	// Unmarshal the decrypted JSON data into an interface.
	log.Info("🧩 Unmarshaling decrypted data")
	var decryptedData interface{}
	err = json.Unmarshal(plaintext, &decryptedData)
	if err != nil {
		log.Error("❌ JSON unmarshaling failed: " + err.Error())
		return nil, err
	}

	// Log successful decryption.
	log.Success("✅ Data decrypted successfully")
	return decryptedData, nil
}
