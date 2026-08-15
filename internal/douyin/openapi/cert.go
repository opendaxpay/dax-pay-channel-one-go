package openapi

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// CertStore：平台证书缓存（cert_no → 公钥）
type CertStore struct {
	mu         sync.RWMutex
	certs      map[string]*rsa.PublicKey
	latestPub  *rsa.PublicKey // 最近加载的证书公钥(转账/分账敏感字段加密用)
	latestSN   string         // 最近加载的证书序列号(十六进制大写, Douyinpay-Serial 头)
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

// Latest：最近一次加载的平台证书公钥
func (s *CertStore) Latest() *rsa.PublicKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestPub
}

// LatestSerial：最近一次加载的平台证书序列号(十六进制大写)
func (s *CertStore) LatestSerial() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.latestSN
}

// Put：写入缓存
func (s *CertStore) Put(serial string, pub *rsa.PublicKey) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.certs[serial] = pub
	s.latestPub = pub
	s.latestSN = serial
}

// platformCertListResp：GET getPlatformCertificates 响应
type platformCertListResp struct {
	Certificates []struct {
		CertNo             string `json:"cert_no"`
		EncryptCertificate struct {
			Algorithm      string `json:"algorithm"`
			Nonce          string `json:"nonce"`
			AssociatedData string `json:"associated_data"`
			CipherText     string `json:"cipher_text"`
		} `json:"encrypt_certificate"`
	} `json:"certificates"`
}

// LoadPlatformCerts：解密平台证书列表并写入缓存，返回命中 serial 的公钥
func (s *CertStore) LoadPlatformCerts(encryptKey string, body []byte, wantSerial string) (*rsa.PublicKey, error) {
	var list platformCertListResp
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, fmt.Errorf("parse certificates: %w", err)
	}
	var found *rsa.PublicKey
	for _, item := range list.Certificates {
		aad := item.EncryptCertificate.AssociatedData
		plain, err := DecryptAESGCM(
			encryptKey,
			item.EncryptCertificate.Nonce,
			aad,
			item.EncryptCertificate.CipherText,
		)
		if err != nil {
			return nil, fmt.Errorf("decrypt cert %s: %w", item.CertNo, err)
		}
		pemStr := string(plain)
		if !strings.Contains(pemStr, "BEGIN") {
			pemStr = "-----BEGIN CERTIFICATE-----\n" + pemStr + "\n-----END CERTIFICATE-----"
		}
		pub, serial, err := ParseCertificate(pemStr)
		if err != nil {
			return nil, fmt.Errorf("parse cert %s: %w", item.CertNo, err)
		}
		key := item.CertNo
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
