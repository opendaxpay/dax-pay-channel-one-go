package transport

import (
	"encoding/base64"
	"strings"
	"testing"
)

// 测试密钥：恰好 32 字节 UTF-8 字符（AES-256）
const testKey = "0123456789abcdef0123456789abcdef"

func TestNewEncryptor_KeyLength(t *testing.T) {
	cases := []struct {
		name string
		key  string
		ok   bool
	}{
		{"valid 32 bytes", testKey, true},
		{"empty", "", false},
		{"too short 16 bytes", "0123456789abcdef", false},
		{"too long 33 bytes", "0123456789abcdef0123456789abcdef0", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enc, err := NewEncryptor(tc.key)
			if tc.ok {
				if err != nil {
					t.Fatalf("expected ok, got err: %v", err)
				}
				if enc == nil {
					t.Fatal("expected non-nil Encryptor")
				}
			} else {
				if err == nil {
					t.Fatal("expected err, got nil")
				}
			}
		})
	}
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	enc, _ := NewEncryptor(testKey)
	cases := []string{
		"",
		"hello",
		`{"outTradeNo":"123456","amount":100}`,
		strings.Repeat("x", 4096),
	}
	for _, plain := range cases {
		ciphertext, err := enc.Encrypt(plain)
		if err != nil {
			t.Fatalf("Encrypt(%q) err: %v", plain, err)
		}
		got, err := enc.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("Decrypt err: %v", err)
		}
		if got != plain {
			t.Fatalf("round-trip mismatch: got %q, want %q", got, plain)
		}
	}
}

func TestEncrypt_IVRandomness(t *testing.T) {
	// 同明文两次加密密文应不同（IV 随机）
	enc, _ := NewEncryptor(testKey)
	c1, _ := enc.Encrypt("same payload")
	c2, _ := enc.Encrypt("same payload")
	if c1 == c2 {
		t.Fatal("expected different ciphertexts due to random IV")
	}
}

func TestEncrypt_CiphertextFormat(t *testing.T) {
	enc, _ := NewEncryptor(testKey)
	ciphertext, _ := enc.Encrypt("test payload")
	combined, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		t.Fatalf("base64 decode err: %v", err)
	}
	// 格式：IV(12) ‖ ciphertext ‖ tag(16)，总长 > 12 + 16
	if len(combined) <= GCMIVLength+16 {
		t.Fatalf("ciphertext too short (missing IV/tag): %d bytes", len(combined))
	}
}

func TestDecrypt_TamperDetection(t *testing.T) {
	enc, _ := NewEncryptor(testKey)
	ciphertext, _ := enc.Encrypt("sensitive data")
	combined, _ := base64.StdEncoding.DecodeString(ciphertext)
	// 篡改密文区一个字节（IV 之后），GCM tag 校验应失败
	combined[GCMIVLength+1] ^= 0xFF
	tampered := base64.StdEncoding.EncodeToString(combined)
	_, err := enc.Decrypt(tampered)
	if err == nil {
		t.Fatal("expected decrypt error on tampered ciphertext")
	}
	if err != ErrDecryptFailed {
		t.Fatalf("expected ErrDecryptFailed, got %v", err)
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	enc, _ := NewEncryptor(testKey)
	_, err := enc.Decrypt("!!!not valid base64!!!")
	if err == nil {
		t.Fatal("expected error on invalid base64")
	}
}

func TestDecrypt_ShortCiphertext(t *testing.T) {
	enc, _ := NewEncryptor(testKey)
	// 只有 12 字节（<= GCMIV_LENGTH），应返回 ErrCiphertextInvalid
	short := base64.StdEncoding.EncodeToString([]byte("123456789012"))
	_, err := enc.Decrypt(short)
	if err != ErrCiphertextInvalid {
		t.Fatalf("expected ErrCiphertextInvalid, got %v", err)
	}
}
