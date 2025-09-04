package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

func DeriveKey(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 10000, 32, sha256.New)
}

func Encrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func Decrypt(data, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func HashPassword(password string) (string, string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}

	hash := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(hash), base64.StdEncoding.EncodeToString(salt), nil
}

func VerifyPassword(password, storedHash, storedSalt string) bool {
	salt, err := base64.StdEncoding.DecodeString(storedSalt)
	if err != nil {
		return false
	}

	hash := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(hash) == storedHash
}

func SimpleEncrypt(data []byte, password string) ([]byte, error) {
	key := sha256.Sum256([]byte(password))
	return Encrypt(data, key[:])
}

func SimpleDecrypt(data []byte, password string) ([]byte, error) {
	key := sha256.Sum256([]byte(password))
	return Decrypt(data, key[:])
}
