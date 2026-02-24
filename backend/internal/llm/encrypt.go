package llm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptAPIKey 用 AES-256-GCM 加密 API Key
// encKey 必须是 32 字节（256 bit），从环境变量 LLM_ENCRYPT_KEY 读取
func EncryptAPIKey(plaintext, encKey string) (string, error) {
	if len(encKey) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes, got %d", len(encKey))
	}
	block, err := aes.NewCipher([]byte(encKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAPIKey 解密 API Key
func DecryptAPIKey(cipherB64, encKey string) (string, error) {
	if len(encKey) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes")
	}
	data, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher([]byte(encKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// APIKeyPrefix 返回 API Key 的可展示前缀
func APIKeyPrefix(plaintext string) string {
	if len(plaintext) <= 10 {
		return plaintext
	}
	return plaintext[:10] + "..."
}
