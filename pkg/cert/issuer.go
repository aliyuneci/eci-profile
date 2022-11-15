package cert

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/keyutil"
)

const (
	CertificateBlockType = "CERTIFICATE"
	MaxCertValidateTime  = time.Hour * 24 * 365 * 100 // ten year self-signed certs
)

type Issuer struct {
	caCert *x509.Certificate
	caKey  crypto.PrivateKey
	caData []byte
}

func NewIssuer(caCertPemData, caKeyPemData []byte) (*Issuer, error) {
	m := &Issuer{}

	m.caData = caCertPemData

	caBlock, _ := pem.Decode(caCertPemData)
	if caBlock == nil {
		return nil, errors.New("decode pem cert failed")
	}

	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return nil, errors.New("parse cert failed")
	}

	if err != nil {
		return nil, errors.Wrap(err, "load ca cert file failed")
	}
	m.caCert = caCert

	block, _ := pem.Decode(caKeyPemData)
	if block == nil {
		return nil, errors.New("decode pem key failed")
	}

	caKey, err := m.parsePrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "parse private key failed")
	}
	m.caKey = caKey

	return m, nil
}

func (m *Issuer) IssueCSR(commonName string, hosts []string) ([]byte, []byte, error) {
	return m.makeCSR(m.caCert, m.caKey, commonName, nil, hosts)
}

func (m *Issuer) GetCAData() []byte {
	return m.caData
}

// Attempt to parse the given private key DER block. OpenSSL 0.9.8 generates
// PKCS #1 private keys by default, while OpenSSL 1.0.0 generates PKCS #8 keys.
// OpenSSL ecparam generates SEC1 EC private keys for ECDSA. We try all three.
func (m *Issuer) parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey, ed25519.PrivateKey:
			return key, nil
		default:
			return nil, errors.New("tls: found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, errors.New("tls: failed to parse private key")
}

func (m *Issuer) makeCSR(caCert *x509.Certificate, caKey crypto.PrivateKey, commonName string, ips, hosts []string) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, errors.Wrap(err, "generate private key failed")
	}

	validFrom := time.Now().Add(-time.Hour) // valid an hour earlier to avoid flakes due to clock skew

	template := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             validFrom,
		NotAfter:              validFrom.Add(MaxCertValidateTime),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	for _, ip := range ips {
		if ip := net.ParseIP(ip); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		}
	}
	template.DNSNames = append(template.DNSNames, hosts...)

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, caCert, &key.PublicKey, caKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create cert failed")
	}

	certBuffer := bytes.Buffer{}
	if err := pem.Encode(&certBuffer, &pem.Block{Type: CertificateBlockType, Bytes: derBytes}); err != nil {
		return nil, nil, err
	}

	// Generate key
	keyBuffer := bytes.Buffer{}
	if err := pem.Encode(&keyBuffer, &pem.Block{Type: keyutil.RSAPrivateKeyBlockType, Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		return nil, nil, err
	}

	return certBuffer.Bytes(), keyBuffer.Bytes(), nil
}
