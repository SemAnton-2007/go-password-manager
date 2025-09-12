// Package crypto предоставляет криптографические функции для менеджера паролей.
//
// Включает:
// - Шифрование и дешифрование данных с использованием AES-GCM
// - Хеширование паролей с PBKDF2 и солью
// - Верификацию паролей
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

// DeriveKey создает cryptographic key из пароля и соли
//
// Parameters:
//
//	password - исходный пароль
//	salt - соль для усиления security
//
// Returns:
//
//	[]byte - derived key длиной 32 байта
//
// Example:
//
//	key := DeriveKey([]byte("password"), []byte("salt"))
func DeriveKey(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 10000, 32, sha256.New)
}

// Encrypt шифрует данные
//
// Parameters:
//
//	data - данные для шифрования
//	key - cryptographic key длиной 32 байта
//
// Returns:
//
//	[]byte - зашифрованные данные с nonce
//	error - ошибка если ключ невалиден или шифрование не удалось
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

// Decrypt дешифрует данные
//
// Parameters:
//
//	data - зашифрованные данные (должны включать nonce)
//	key - cryptographic key использованный для шифрования
//
// Returns:
//
//	[]byte - расшифрованные данные
//	error - ошибка если данные повреждены или ключ неверный
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

// HashPassword создает безопасный хэш пароля
//
// Parameters:
//
//	password - пароль для хэширования
//
// Returns:
//
//	string - base64-encoded хэш пароля
//	string - base64-encoded соль
//	error - ошибка генерации
func HashPassword(password string) (string, string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}

	hash := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(hash), base64.StdEncoding.EncodeToString(salt), nil
}

// VerifyPassword проверяет пароль
//
// Parameters:
//
//	password - пароль для проверки
//	storedHash - stored хэш пароля (base64)
//	storedSalt - stored соль (base64)
//
// Returns:
//
//	bool - true если пароль верный
func VerifyPassword(password, storedHash, storedSalt string) bool {
	salt, err := base64.StdEncoding.DecodeString(storedSalt)
	if err != nil {
		return false
	}

	hash := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	return base64.StdEncoding.EncodeToString(hash) == storedHash
}
