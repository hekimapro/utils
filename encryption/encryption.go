package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/hekimapro/utils/log"
	"github.com/hekimapro/utils/models"
)

// pad applies PKCS7 padding to the plaintext
func pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padText...)
}

// unpad removes PKCS7 padding from the decrypted plaintext
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

// Encrypt encrypts data using AES in CBC mode and returns an encoded payload
func Encrypt(data interface{}, encryptionKey, initializationVector, encryptionType string) (models.EncryptReturnType, error) {
	log.Info("Starting encryption process")

	if encryptionKey == "" {
		log.Error("Encryption key is empty")
		return models.EncryptReturnType{}, errors.New("encryption key must be a non-empty string")
	}

	if len(initializationVector) != aes.BlockSize {
		log.Error("Invalid initialization vector length")
		return models.EncryptReturnType{}, errors.New("initialization vector must be exactly 16 bytes long")
	}

	if encryptionType != "base64" && encryptionType != "hex" {
		log.Error("Unsupported encryption type provided")
		return models.EncryptReturnType{}, errors.New("invalid encryption type (use 'base64' or 'hex')")
	}

	dataToEncrypt, err := json.Marshal(data)
	if err != nil {
		log.Error("Failed to marshal input data: " + err.Error())
		return models.EncryptReturnType{}, err
	}

	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		log.Error("Failed to create AES cipher: " + err.Error())
		return models.EncryptReturnType{}, err
	}

	paddedData := pad(dataToEncrypt, aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, []byte(initializationVector))
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	var encryptedPayload string
	if encryptionType == "base64" {
		encryptedPayload = base64.StdEncoding.EncodeToString(ciphertext)
	} else {
		encryptedPayload = hex.EncodeToString(ciphertext)
	}

	log.Success("Data successfully encrypted")
	return models.EncryptReturnType{Payload: encryptedPayload}, nil
}

// Decrypt decrypts AES-encrypted data in CBC mode and returns original data
func Decrypt(encryptedData models.EncryptReturnType, encryptionKey, initializationVector, encryptionType string) (interface{}, error) {
	log.Info("Starting decryption process")

	if encryptionKey == "" {
		log.Error("Encryption key is empty")
		return nil, errors.New("encryption key must be a non-empty string")
	}

	if len(initializationVector) != aes.BlockSize {
		log.Error("Invalid initialization vector length")
		return nil, errors.New("initialization vector must be exactly 16 bytes long")
	}

	if encryptionType != "base64" && encryptionType != "hex" {
		log.Error("Unsupported encryption type provided")
		return nil, errors.New("invalid encryption type")
	}

	var ciphertext []byte
	var err error
	if encryptionType == "base64" {
		ciphertext, err = base64.StdEncoding.DecodeString(encryptedData.Payload)
	} else {
		ciphertext, err = hex.DecodeString(encryptedData.Payload)
	}
	if err != nil {
		log.Error("Failed to decode encrypted payload: " + err.Error())
		return nil, err
	}

	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		log.Error("Failed to create AES cipher: " + err.Error())
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, []byte(initializationVector))
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	plaintext, err = unpad(plaintext)
	if err != nil {
		log.Error("Failed to unpad decrypted data: " + err.Error())
		return nil, err
	}

	var decryptedData interface{}
	err = json.Unmarshal(plaintext, &decryptedData)
	if err != nil {
		log.Error("Failed to unmarshal decrypted data: " + err.Error())
		return nil, err
	}

	log.Success("Data successfully decrypted")
	return decryptedData, nil
}
