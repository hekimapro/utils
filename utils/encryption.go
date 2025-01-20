package utils

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

// pad applies PKCS7 padding to the plaintext.
func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

// unpad removes PKCS7 padding from the plaintext.
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

// Encrypts the provided data using the specified parameters.
func Encrypt(data interface{}, encryptionKey, initializationVector, encryptionType string) (models.EncryptReturnType, error) {

	// Validate parameters
	if encryptionKey == "" {
		return models.EncryptReturnType{}, errors.New("encryption key must be a non-empty string")
	}
	if len(initializationVector) != aes.BlockSize {
		return models.EncryptReturnType{}, errors.New("initialization vector must be exactly 16 bytes long")
	}
	if encryptionType != "base64" && encryptionType != "hex" {
		return models.EncryptReturnType{}, errors.New("invalid encryption type (base64 || hex)")
	}

	// Convert data to a string
	dataToEncrypt, err := json.Marshal(data)
	if err != nil {
		return models.EncryptReturnType{}, err
	}

	// Create cipher block
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return models.EncryptReturnType{}, err
	}

	// Pad data to block size
	paddedData := pad(dataToEncrypt, aes.BlockSize)

	// Create CBC mode encrypter
	mode := cipher.NewCBCEncrypter(block, []byte(initializationVector))

	// Encrypt data
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	// Convert encrypted data to the specified encryption type
	var encryptedPayload string
	if encryptionType == "base64" {
		encryptedPayload = base64.StdEncoding.EncodeToString(ciphertext)
	} else {
		encryptedPayload = hex.EncodeToString(ciphertext)
	}

	return models.EncryptReturnType{Payload: encryptedPayload}, nil
}

// Decrypts the provided encrypted data using the specified parameters.
func Decrypt(encryptedData models.EncryptReturnType, encryptionKey, initializationVector, encryptionType string) (interface{}, error) {

	// Validate parameters
	if encryptionKey == "" {
		return nil, errors.New("encryption key must be a non-empty string")
	}
	if len(initializationVector) != aes.BlockSize {
		return nil, errors.New("initialization vector must be exactly 16 bytes long")
	}
	if encryptionType != "base64" && encryptionType != "hex" {
		return nil, errors.New("invalid encryption type")
	}

	// Decode encrypted data from the specified encryption type
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

	// Create cipher block
	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return nil, err
	}

	// Create CBC mode decrypter
	mode := cipher.NewCBCDecrypter(block, []byte(initializationVector))

	// Decrypt data
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// Unpad data
	plaintext, err = unpad(plaintext)
	if err != nil {
		return nil, err
	}

	// Convert decrypted data to original format
	var decryptedData interface{}
	err = json.Unmarshal(plaintext, &decryptedData)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
