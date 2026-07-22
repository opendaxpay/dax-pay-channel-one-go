package openapi

import (
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// CertSN：单本证书 SN（对标 Java AntCertificationUtil.getCertSN）
func CertSN(certPEM string) (string, error) {
	cert, err := parseFirstCert(certPEM)
	if err != nil {
		return "", err
	}
	return certSNFromX509(cert), nil
}

// RootCertSN：根证书 SN（多本用 _ 拼接）
//
// 对标支付宝 Java SDK：根证书 PEM 常含多本证书，其中可能有 Go 无法解析的椭圆曲线证书；
// 解析失败的跳过，仅对 SHA256WithRSA 签名的证书计算 SN。
func RootCertSN(rootCertPEM string) (string, error) {
	certs, err := parseAllCerts(rootCertPEM, true)
	if err != nil {
		return "", err
	}
	var sns []string
	for _, c := range certs {
		if c.SignatureAlgorithm != x509.SHA256WithRSA {
			continue
		}
		sns = append(sns, certSNFromX509(c))
	}
	if len(sns) == 0 {
		return "", errors.New("no SHA256WithRSA cert in root cert content")
	}
	return strings.Join(sns, "_"), nil
}

// ExtractRSAPublicKey：从支付宝公钥证书提取 RSA 公钥
func ExtractRSAPublicKey(certPEM string) (*rsa.PublicKey, error) {
	cert, err := parseFirstCert(certPEM)
	if err != nil {
		return nil, err
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("cert public key is not RSA")
	}
	return pub, nil
}

func certSNFromX509(cert *x509.Certificate) string {
	issuer := cert.Issuer.String()
	serial := cert.SerialNumber.String()
	sum := md5.Sum([]byte(issuer + serial))
	return strings.ToLower(hex.EncodeToString(sum[:]))
}

func parseFirstCert(pemOrContent string) (*x509.Certificate, error) {
	// 应用/支付宝公钥证书也可能夹杂无法解析的块，跳过无效后取第一本成功解析的
	certs, err := parseAllCerts(pemOrContent, true)
	if err != nil {
		return nil, err
	}
	return certs[0], nil
}

// parseAllCerts：解析 PEM 中全部证书；skipInvalid=true 时跳过无法解析的块（如 unsupported elliptic curve）
func parseAllCerts(content string, skipInvalid bool) ([]*x509.Certificate, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("empty cert content")
	}
	var certs []*x509.Certificate
	rest := []byte(content)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			if skipInvalid {
				// 支付宝根证书包含 SM2/特殊曲线等证书，Go 标准库会报 unsupported elliptic curve
				continue
			}
			return nil, fmt.Errorf("parse cert: %w", err)
		}
		certs = append(certs, c)
	}
	if len(certs) == 0 {
		return nil, errors.New("no PEM certificate found")
	}
	return certs, nil
}
