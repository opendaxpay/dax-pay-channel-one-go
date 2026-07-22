package openapi_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"strings"
)

func rsaMarshalPKCS8(priv *rsa.PrivateKey) ([]byte, error) {
	return x509.MarshalPKCS8PrivateKey(priv)
}

func encodeStd(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func chunk64(der []byte) string {
	s := base64.StdEncoding.EncodeToString(der)
	var b strings.Builder
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s[i:end])
	}
	return b.String()
}
