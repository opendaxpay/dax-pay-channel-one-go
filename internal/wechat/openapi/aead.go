package openapi

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// DecryptAEAD：AEAD_AES_256_GCM 解密（apiKeyV3 为 32 字节密钥）
//
// ciphertextBase64 为微信返回的 Base64（含 auth tag 后缀）。
func DecryptAEAD(apiKeyV3, associatedData, nonce, ciphertextBase64 string) ([]byte, error) {
	key := []byte(apiKeyV3)
	if len(key) != 32 {
		return nil, fmt.Errorf("apiKeyV3 must be 32 bytes, got %d", len(key))
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, fmt.Errorf("decode ciphertext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("nonce length %d != %d", len(nonce), gcm.NonceSize())
	}
	plain, err := gcm.Open(nil, []byte(nonce), ciphertext, []byte(associatedData))
	if err != nil {
		return nil, fmt.Errorf("aead open: %w", err)
	}
	return plain, nil
}

// EncryptAEAD：AEAD_AES_256_GCM 加密（测试用），返回 Base64 ciphertext（含 tag）
func EncryptAEAD(apiKeyV3, associatedData, nonce string, plaintext []byte) (string, error) {
	key := []byte(apiKeyV3)
	if len(key) != 32 {
		return "", fmt.Errorf("apiKeyV3 must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(nonce) != gcm.NonceSize() {
		return "", fmt.Errorf("nonce length %d != %d", len(nonce), gcm.NonceSize())
	}
	out := gcm.Seal(nil, []byte(nonce), plaintext, []byte(associatedData))
	return base64.StdEncoding.EncodeToString(out), nil
}
