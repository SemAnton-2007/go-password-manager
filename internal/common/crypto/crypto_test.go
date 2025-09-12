package crypto

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestDeriveKey(t *testing.T) {
	password := []byte("testpassword")
	salt := []byte("testsalt")

	key := DeriveKey(password, salt)

	if len(key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key))
	}

	// Проверяем, что тот же пароль и соль дают тот же ключ
	key2 := DeriveKey(password, salt)
	if string(key) != string(key2) {
		t.Error("Same password and salt should produce same key")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key := sha256.Sum256([]byte("testpassword"))
	testData := []byte("Hello, World! This is a test message.")

	encrypted, err := Encrypt(testData, key[:])
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if len(encrypted) <= len(testData) {
		t.Error("Encrypted data should be longer than plaintext")
	}

	decrypted, err := Decrypt(encrypted, key[:])
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != string(testData) {
		t.Errorf("Decrypted data doesn't match original. Got: %s, Expected: %s",
			string(decrypted), string(testData))
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"

	hash, salt, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash should not be empty")
	}
	if salt == "" {
		t.Error("Salt should not be empty")
	}

	// Проверяем base64 encoding
	_, err = base64.StdEncoding.DecodeString(hash)
	if err != nil {
		t.Errorf("Hash is not valid base64: %v", err)
	}

	_, err = base64.StdEncoding.DecodeString(salt)
	if err != nil {
		t.Errorf("Salt is not valid base64: %v", err)
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"

	hash, salt, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Правильный пароль
	if !VerifyPassword(password, hash, salt) {
		t.Error("Correct password should verify successfully")
	}

	// Неправильный пароль
	if VerifyPassword("wrongpassword", hash, salt) {
		t.Error("Wrong password should not verify")
	}
}

func TestEncryptDecryptEmptyData(t *testing.T) {
	key := sha256.Sum256([]byte("testpassword"))

	// Пустые данные
	emptyData := []byte{}
	encrypted, err := Encrypt(emptyData, key[:])
	if err != nil {
		t.Fatalf("Encrypt empty data failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key[:])
	if err != nil {
		t.Fatalf("Decrypt empty data failed: %v", err)
	}

	if len(decrypted) != 0 {
		t.Errorf("Expected empty data, got %d bytes", len(decrypted))
	}
}

func TestDecryptCorruptedData(t *testing.T) {
	key := sha256.Sum256([]byte("testpassword"))
	corruptedData := []byte("corrupted encrypted data")

	_, err := Decrypt(corruptedData, key[:])
	if err == nil {
		t.Error("Should fail with corrupted data")
	}
}

func TestEncryptDecryptLargeData(t *testing.T) {
	// Большие данные
	largeData := make([]byte, 1024*10) // 10KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	key := sha256.Sum256([]byte("testpassword"))

	encrypted, err := Encrypt(largeData, key[:])
	if err != nil {
		t.Fatalf("Encrypt large data failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key[:])
	if err != nil {
		t.Fatalf("Decrypt large data failed: %v", err)
	}

	if len(decrypted) != len(largeData) {
		t.Errorf("Length mismatch. Got: %d, Expected: %d", len(decrypted), len(largeData))
	}

	for i := range largeData {
		if decrypted[i] != largeData[i] {
			t.Errorf("Data mismatch at index %d", i)
			break
		}
	}
}

func TestEncryptWithShortKey(t *testing.T) {
	shortKey := []byte("short") // меньше чем требуется для AES
	testData := []byte("test data")

	_, err := Encrypt(testData, shortKey)
	if err == nil {
		t.Error("Should fail with short key")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	testData := []byte("test data")
	key1 := sha256.Sum256([]byte("password1"))
	key2 := sha256.Sum256([]byte("password2"))

	encrypted, err := Encrypt(testData, key1[:])
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(encrypted, key2[:])
	if err == nil {
		t.Error("Should fail with wrong key")
	}
}

func TestHashPasswordEmpty(t *testing.T) {
	_, _, err := HashPassword("")
	if err != nil {
		t.Errorf("HashPassword with empty password failed: %v", err)
	}
}

func TestVerifyPasswordEmpty(t *testing.T) {
	// Пустые хэш и соль
	result := VerifyPassword("test", "", "")
	if result {
		t.Error("VerifyPassword should fail with empty hash and salt")
	}

	// Невалидный base64
	result = VerifyPassword("test", "invalid-base64!", "invalid-base64!")
	if result {
		t.Error("VerifyPassword should fail with invalid base64")
	}
}
