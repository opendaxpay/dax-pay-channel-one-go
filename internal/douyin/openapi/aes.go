package openapi

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
)

// DecryptAESGCM：AEAD-AES-256-GCM 解密（encryptKey 须正好 32 字节）
//
// ciphertextB64 为 Base64（含 GCM auth tag 后缀）；associatedData 可空。
func DecryptAESGCM(key, nonce, associatedData, ciphertextB64 string) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryptKey must be 32 bytes, got %d", len(key))
	}
	ct, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher([]byte(key))
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
	var aad []byte
	if associatedData != "" {
		aad = []byte(associatedData)
	}
	return gcm.Open(nil, []byte(nonce), ct, aad)
}

// EncryptAESGCM：AEAD-AES-256-GCM 加密（单测用），返回 Base64 ciphertext（含 tag）
func EncryptAESGCM(key, nonce, associatedData string, plaintext []byte) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("encryptKey must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher([]byte(key))
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
	var aad []byte
	if associatedData != "" {
		aad = []byte(associatedData)
	}
	out := gcm.Seal(nil, []byte(nonce), plaintext, aad)
	return base64.StdEncoding.EncodeToString(out), nil
}
