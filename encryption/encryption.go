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

// pad applies PKCS7 padding to the plaintext to ensure the data length is a multiple of the block size.
func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

// unpad removes PKCS7 padding from the decrypted plaintext.
func unpad(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, errors.New("invalid padding size")
	}
	padding := int(src[length-1])
	if padding > length || padding == 0 {
		return nil, errors.New("invalid padding size")
	}
	return src[:length-padding], nil
}

// Encrypt encrypts the provided data using AES encryption in CBC mode with the specified parameters.
// encryptionKey: The key used for encryption (must be 16, 24, or 32 bytes).
// initializationVector: The IV used for CBC mode (must be 16 bytes).
// encryptionType: The desired encryption format ("base64" or "hex").
func Encrypt(data interface{}, encryptionKey, initializationVector, encryptionType string) (models.EncryptReturnType, error) {

	// Validate encryption parameters
	if encryptionKey == "" {
		return models.EncryptReturnType{}, errors.New("encryption key must be a non-empty string")
	}
	if len(initializationVector) != aes.BlockSize {
		return models.EncryptReturnType{}, errors.New("initialization vector must be exactly 16 bytes long")
	}
	if encryptionType != "base64" && encryptionType != "hex" {
		return models.EncryptReturnType{}, errors.New("invalid encryption type (use 'base64' or 'hex')")
	}

	// Convert data to JSON format
	dataToEncrypt, err := json.Marshal(data)
	if err != nil {
		return models.EncryptReturnType{}, err
	}

	// Create AES cipher block
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return models.EncryptReturnType{}, err
	}

	// Pad the data to match block size
	paddedData := pad(dataToEncrypt, aes.BlockSize)

	// Initialize CBC encrypter
	mode := cipher.NewCBCEncrypter(block, []byte(initializationVector))

	// Encrypt data
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	// Encode encrypted data based on specified format
	var encryptedPayload string
	if encryptionType == "base64" {
		encryptedPayload = base64.StdEncoding.EncodeToString(ciphertext)
	} else {
		encryptedPayload = hex.EncodeToString(ciphertext)
	}

	return models.EncryptReturnType{Payload: encryptedPayload}, nil
}

// Decrypt decrypts the provided encrypted data using AES encryption in CBC mode with the specified parameters.
// encryptedData: The encrypted data in the specified format ("base64" or "hex").
// encryptionKey: The key used for decryption (must be 16, 24, or 32 bytes).
// initializationVector: The IV used for CBC mode (must be 16 bytes).
// encryptionType: The encryption format ("base64" or "hex") that was used for encryption.
func Decrypt(encryptedData models.EncryptReturnType, encryptionKey, initializationVector, encryptionType string) (interface{}, error) {

	// Validate decryption parameters
	if encryptionKey == "" {
		return nil, errors.New("encryption key must be a non-empty string")
	}
	if len(initializationVector) != aes.BlockSize {
		return nil, errors.New("initialization vector must be exactly 16 bytes long")
	}
	if encryptionType != "base64" && encryptionType != "hex" {
		return nil, errors.New("invalid encryption type")
	}

	// Decode encrypted data based on the specified format
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

	// Create AES cipher block for decryption
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	// Initialize CBC decrypter
	mode := cipher.NewCBCDecrypter(block, []byte(initializationVector))

	// Decrypt data
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Remove padding from decrypted data
	plaintext, err = unpad(plaintext)
	if err != nil {
		return nil, err
	}

	// Convert decrypted data back to original format (interface)
	var decryptedData interface{}
	err = json.Unmarshal(plaintext, &decryptedData)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
