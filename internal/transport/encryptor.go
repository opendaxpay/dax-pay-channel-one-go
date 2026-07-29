// Package transport：主应用 ↔ 通道子应用 报文透明加密层
//
// 对齐 Java 端 `cn.daxpay.open.platform.common.util.encrypt.ChannelAesGcmEncryptor`：
//   - 算法：AES-256-GCM（无额外 HMAC，GCM tag 自带完整性校验）
//   - 密钥：恰好 32 字节 UTF-8 字符（AES-256）
//   - IV：12 字节随机，每次加密新生成
//   - GCM Tag：128 位（16 字节，附在密文末尾）
//   - 密文格式：Base64( IV(12) ‖ ciphertext ‖ tag(16) )
//
// Go 的 `cipher.NewGCM(block)` 默认 tag 长度 16 字节，`aead.Seal(nil, iv, plain, nil)`
// 输出 `ciphertext ‖ tag`，与 Java `Cipher.doFinal()` 输出顺序一致，密文可互通。
package transport

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const (
	// KeyLength：AES-256 密钥字节数（恰好 32 个 UTF-8 字符）
	KeyLength = 32
	// GCMIVLength：GCM IV 字节长度
	GCMIVLength = 12
)

// 与 Java ChannelAesGcmEncryptor messageKey 对齐（由中间件层用 i18n.T 解析）
var (
	// ErrEncryptFailed 通道传输加密失败
	ErrEncryptFailed = errors.New("channel.error.transportEncrypt.encryptFailed")
	// ErrCiphertextInvalid 通道传输密文长度非法
	ErrCiphertextInvalid = errors.New("channel.error.transportEncrypt.ciphertextInvalid")
	// ErrDecryptFailed 通道传输解密失败
	ErrDecryptFailed = errors.New("channel.error.transportEncrypt.decryptFailed")
)

// Encryptor：通道传输 AES-256-GCM 加解密器（线程安全，无内部状态）
type Encryptor struct {
	aead cipher.AEAD
}

// NewEncryptor 构造加解密器。
//
// key 须恰好 32 字节 UTF-8 字符（AES-256）。密钥长度非法时返回固定可读 error
// （启动期校验，非 i18n key，由 main.go 启动失败日志直接展示）。
func NewEncryptor(key string) (*Encryptor, error) {
	if len(key) != KeyLength {
		return nil, fmt.Errorf("channel transport key must be %d bytes (got %d)", KeyLength, len(key))
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Encryptor{aead: aead}, nil
}

// Encrypt 加密明文，返回 Base64(IV‖ciphertext‖tag)。
// 空明文允许加密（IV + 空 ciphertext + tag），与 Java 行为一致。
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	iv := make([]byte, GCMIVLength)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", ErrEncryptFailed
	}
	// Seal(dst, nonce, plaintext, additionalData)：输出 ciphertext‖tag
	sealed := e.aead.Seal(nil, iv, []byte(plaintext), nil)
	combined := make([]byte, 0, len(iv)+len(sealed))
	combined = append(combined, iv...)
	combined = append(combined, sealed...)
	return base64.StdEncoding.EncodeToString(combined), nil
}

// Decrypt 解密 Base64(IV‖ciphertext‖tag)。
// IV 长度非法、Base64 解码失败、GCM 认证失败均返回对应的 i18n key error。
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	combined, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrDecryptFailed
	}
	if len(combined) <= GCMIVLength {
		return "", ErrCiphertextInvalid
	}
	// Open(dst, nonce, ciphertext‖tag, additionalData)：认证并解密
	plain, err := e.aead.Open(nil, combined[:GCMIVLength], combined[GCMIVLength:], nil)
	if err != nil {
		return "", ErrDecryptFailed
	}
	return string(plain), nil
}
