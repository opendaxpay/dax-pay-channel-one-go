package openapi

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"sync"
)

// CertStore：平台证书缓存（按 serial → 公钥）
type CertStore struct {
	mu     sync.RWMutex
	certs  map[string]*rsa.PublicKey
	latest *rsa.PublicKey // 最近一次加载的证书(转账敏感字段加密用, 微信一般只有一张有效证书)
}

// NewCertStore：新建空缓存
func NewCertStore() *CertStore {
	return &CertStore{certs: make(map[string]*rsa.PublicKey)}
}

// Get：按序列号取公钥
func (s *CertStore) Get(serial string) *rsa.PublicKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.certs[serial]
}

// Latest：最近一次加载的平台证书公钥(证书轮换场景取最新)
func (s *CertStore) Latest() *rsa.PublicKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latest
}

// Put：写入缓存
func (s *CertStore) Put(serial string, pub *rsa.PublicKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.certs[serial] = pub
	s.latest = pub
}

// platformCertListResp：GET /v3/certificates 响应
type platformCertListResp struct {
	Data []struct {
		SerialNo           string `json:"serial_no"`
		EncryptCertificate struct {
			Algorithm      string `json:"algorithm"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
			Ciphertext     string `json:"ciphertext"`
		} `json:"encrypt_certificate"`
	} `json:"data"`
}

// LoadPlatformCerts：解密平台证书列表并写入缓存，返回命中 serial 的公钥
func (s *CertStore) LoadPlatformCerts(apiKeyV3 string, body []byte, wantSerial string) (*rsa.PublicKey, error) {
	var list platformCertListResp
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse certificates: %w", err)
	}
	var found *rsa.PublicKey
	for _, item := range list.Data {
		plain, err := DecryptAEAD(
			apiKeyV3,
			item.EncryptCertificate.AssociatedData,
			item.EncryptCertificate.Nonce,
			item.EncryptCertificate.Ciphertext,
		)
		if err != nil {
			return nil, fmt.Errorf("decrypt cert %s: %w", item.SerialNo, err)
		}
		pub, serial, err := ParseCertificate(string(plain))
		if err != nil {
			return nil, fmt.Errorf("parse cert %s: %w", item.SerialNo, err)
		}
		// 优先用响应里的 serial_no；解析出的序列号作兜底
		key := item.SerialNo
		if key == "" {
			key = serial
		}
		s.Put(key, pub)
		if wantSerial != "" && (key == wantSerial || serial == wantSerial) {
			found = pub
		}
	}
	if wantSerial != "" && found == nil {
		found = s.Get(wantSerial)
	}
	return found, nil
}
