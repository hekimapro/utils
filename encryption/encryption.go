package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/hekimapro/utils/models"
)

// pad applies PKCS7 padding to the plaintext
// Ensures the data length is a multiple of the AES block size by adding padding bytes
func pad(src []byte, blockSize int) []byte {
	// Calculate the number of padding bytes needed
	padding := blockSize - len(src)%blockSize
	// Create padding bytes with the padding length value
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	// Append padding to the source data
	return append(src, padText...)
}

// unpad removes PKCS7 padding from the decrypted plaintext
// Validates the padding and returns the unpadded data or an error if invalid
func unpad(src []byte) ([]byte, error) {
	// Check if the input is empty
	length := len(src)
	if length == 0 {
		return nil, errors.New("invalid padding size")
	}
	// Extract the padding length from the last byte
	padding := int(src[length-1])
	// Validate the padding length
	if padding > length || padding == 0 {
		return nil, errors.New("invalid padding size")
	}
	// Return the data without the padding bytes
	return src[:length-padding], nil
}

// Encrypt encrypts data using AES in CBC mode
// Converts input data to JSON, pads it, encrypts it, and encodes the result
// Supports base64 or hex encoding based on encryptionType
func Encrypt(data interface{}, encryptionKey, initializationVector, encryptionType string) (models.EncryptReturnType, error) {
	// Validate the encryption key is non-empty
	if encryptionKey == "" {
		return models.EncryptReturnType{}, errors.New("encryption key must be a non-empty string")
	}
	// Validate the initialization vector length (must be 16 bytes for AES)
	if len(initializationVector) != aes.BlockSize {
		return models.EncryptReturnType{}, errors.New("initialization vector must be exactly 16 bytes long")
	}
	// Validate the encryption type (must be "base64" or "hex")
	if encryptionType != "base64" && encryptionType != "hex" {
		return models.EncryptReturnType{}, errors.New("invalid encryption type (use 'base64' or 'hex')")
	}

	// Marshal the input data to JSON format
	dataToEncrypt, err := json.Marshal(data)
	if err != nil {
		return models.EncryptReturnType{}, err
	}

	// Create an AES cipher block with the provided key
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return models.EncryptReturnType{}, err
	}

	// Apply PKCS7 padding to the JSON data
	paddedData := pad(dataToEncrypt, aes.BlockSize)

	// Initialize CBC mode encrypter with the block and IV
	mode := cipher.NewCBCEncrypter(block, []byte(initializationVector))

	// Encrypt the padded data
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	// Encode the ciphertext based on the specified format
	var encryptedPayload string
	if encryptionType == "base64" {
		encryptedPayload = base64.StdEncoding.EncodeToString(ciphertext)
	} else {
		encryptedPayload = hex.EncodeToString(ciphertext)
	}

	// Return the encrypted payload wrapped in EncryptReturnType
	return models.EncryptReturnType{Payload: encryptedPayload}, nil
}

// Decrypt decrypts AES-encrypted data in CBC mode
// Decodes the input, decrypts it, removes padding, and unmarshals to the original data
// Supports base64 or hex decoding based on encryptionType
func Decrypt(encryptedData models.EncryptReturnType, encryptionKey, initializationVector, encryptionType string) (interface{}, error) {
	// Validate the encryption key is non-empty
	if encryptionKey == "" {
		return nil, errors.New("encryption key must be a non-empty string")
	}
	// Validate the initialization vector length (must be 16 bytes for AES)
	if len(initializationVector) != aes.BlockSize {
		return nil, errors.New("initialization vector must be exactly 16 bytes long")
	}
	// Validate the encryption type (must be "base64" or "hex")
	if encryptionType != "base64" && encryptionType != "hex" {
		return nil, errors.New("invalid encryption type")
	}

	// Decode the encrypted data based on the specified format
	var ciphertext []byte
	var err error
	if encryptionType == "base64" {
		ciphertext, err = base64.StdEncoding.DecodeString(encryptedData.Payload)
	} else {
		ciphertext, err = hex.DecodeString(encryptedData.Payload)
	}
	if err != nil {
		return nil, err
	}

	// Create an AES cipher block with the provided key
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	// Initialize CBC mode decrypter with the block and IV
	mode := cipher.NewCBCDecrypter(block, []byte(initializationVector))

	// Decrypt the ciphertext
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove PKCS7 padding from the decrypted data
	plaintext, err = unpad(plaintext)
	if err != nil {
		return nil, err
	}

	// Unmarshal the decrypted JSON data back to an interface
	var decryptedData interface{}
	err = json.Unmarshal(plaintext, &decryptedData)
	if err != nil {
		return nil, err
	}

	// Return the decrypted data
	return decryptedData, nil
}