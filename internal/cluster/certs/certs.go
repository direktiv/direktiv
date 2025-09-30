package certs

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	rotationInterval = 10
	secretName       = "direktiv-tls-secret" //nolint:gosec
	dummyKey         = "dummy.crt"

	hoursToRefresh = 168
)

type changeMarker struct {
	Time int64 `json:"time"`
}

func (c *CertificateUpdater) requiresRefresh(ctx context.Context) (bool, error) {
	s, err := c.client.CoreV1().Secrets(c.ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	// check if it is still dummy
	_, ok := s.Data[dummyKey]

	// it is still the dummy we need to generate the certs
	if ok {
		return true, nil
	}

	// it has a last change data field
	m := s.Data["marker"]

	var marker changeMarker
	err = json.Unmarshal(m, &marker)
	if err != nil {
		return false, err
	}

	updated := time.Unix(marker.Time, 0)
	if time.Now().After(updated.Add(hoursToRefresh * time.Hour)) {
		return true, nil
	}

	// mark certs ready
	return false, nil
}

func (c *CertificateUpdater) checkAndUpdate(ctx context.Context) error {
	r, err := c.requiresRefresh(ctx)
	if err != nil {
		return err
	}

	if r {
		slog.Info("certificates require refresh")
		err = c.rotateCerts(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *CertificateUpdater) rotateCerts(ctx context.Context) error {
	slog.Info("rotating certificates")
	caTemplate, caPrivKey, err := generateCA()
	if err != nil {
		return err
	}

	srvCrt, srvKey, err := generateCerts(2, c.ns, caTemplate, caPrivKey)
	if err != nil {
		return err
	}

	secret, err := c.client.CoreV1().Secrets(c.ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	caCrt, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	secret.Data = map[string][]byte{}

	marker := &changeMarker{
		Time: time.Now().Unix(),
	}

	b, err := json.Marshal(marker)
	if err != nil {
		return err
	}

	secret.Data["marker"] = b

	type certInfo struct {
		name, block string
		data        []byte
	}

	certs := []certInfo{
		{
			name:  "ca.crt",
			data:  caCrt,
			block: "CERTIFICATE",
		},
		{
			name:  "ca.key",
			data:  x509.MarshalPKCS1PrivateKey(caPrivKey),
			block: "RSA PRIVATE KEY",
		},
		{
			name:  "server.crt",
			data:  srvCrt,
			block: "CERTIFICATE",
		},
		{
			name:  "server.key",
			data:  srvKey,
			block: "RSA PRIVATE KEY",
		},
	}

	for i := range certs {
		h, err := writePem(certs[i].block, certs[i].data)
		if err != nil {
			return err
		}
		secret.Data[certs[i].name] = h
	}

	slog.Info("storing new certificates")
	_, err = c.client.CoreV1().Secrets(c.ns).Update(ctx, secret, metav1.UpdateOptions{})

	return err
}

func generateCA() (x509.Certificate, *rsa.PrivateKey, error) {
	caPriv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return x509.Certificate{}, &rsa.PrivateKey{}, err
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Direktiv CA"},
			CommonName:   "ca.direktiv-nats",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	return caTemplate, caPriv, nil
}

func generateCerts(id int64, ns string, caTemplate x509.Certificate, caPriv *rsa.PrivateKey) ([]byte, []byte, error) {
	serverPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return []byte{}, []byte{}, err
	}

	// set the deployment name in dns names
	deploymentName := os.Getenv("DIREKTIV_DEPLOYMENT_NAME")

	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
	extKeyUsage := []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(id),
		Subject: pkix.Name{
			Organization: []string{"Direktiv CA"},
			CommonName:   fmt.Sprintf("*.%s-nats", deploymentName),
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0),
		KeyUsage:    keyUsage,
		ExtKeyUsage: extKeyUsage,
		DNSNames: []string{
			fmt.Sprintf("*.%s-nats-headless", deploymentName),
			fmt.Sprintf("%s-nats.%s.svc", deploymentName, ns),
		},
	}

	crtBytes, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &caTemplate, &serverPriv.PublicKey, caPriv)

	return crtBytes, x509.MarshalPKCS1PrivateKey(serverPriv), err
}

func writePem(blockType string, in []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := pem.Encode(&buf, &pem.Block{Type: blockType, Bytes: in})

	return buf.Bytes(), err
}
