// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateTestCert creates a self-signed certificate and key and writes them
// to PEM files in the given directory. Returns cert and key file paths.
func generateTestCert(t *testing.T, dir, prefix string) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	serial := big.NewInt(time.Now().UnixNano())
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: prefix + ".example.com",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(1 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("failed to create certificate: %v", err)
	}

	certFile := filepath.Join(dir, prefix+".pem")
	keyFile := filepath.Join(dir, prefix+"-key.pem")

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	return certFile, keyFile
}

// certLoaderSerial extracts the serial number of the first certificate
// returned by the loader's getCertificate callback.
func certLoaderSerial(cl *certificateLoader) *big.Int {
	cert, err := cl.getCertificate(nil)
	if err != nil || cert == nil {
		return nil
	}
	x509Cert, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil
	}
	return x509Cert.SerialNumber
}

// TestCertificateLoaderInitialLoad verifies that a new certificateLoader
// loads the certificate successfully and serves it via getCertificate.
func TestCertificateLoaderInitialLoad(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := generateTestCert(t, dir, "initial")

	cl, err := newCertificateLoader(certFile, keyFile)
	if err != nil {
		t.Fatalf("newCertificateLoader failed: %v", err)
	}

	cert, err := cl.getCertificate(nil)
	if err != nil {
		t.Fatalf("getCertificate failed: %v", err)
	}
	if cert == nil {
		t.Fatal("getCertificate returned nil certificate")
	}
	if len(cert.Certificate) == 0 {
		t.Fatal("certificate chain is empty")
	}
	if certLoaderSerial(cl) == nil {
		t.Fatal("unable to parse certificate serial")
	}
}

// TestCertificateLoaderReload verifies that Reload picks up a new certificate
// written to disk.
func TestCertificateLoaderReload(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := generateTestCert(t, dir, "reload")

	cl, err := newCertificateLoader(certFile, keyFile)
	if err != nil {
		t.Fatalf("newCertificateLoader failed: %v", err)
	}

	origSerial := certLoaderSerial(cl)
	if origSerial == nil {
		t.Fatal("unable to parse original certificate serial")
	}

	// Write a new certificate to the same file paths.
	generateTestCert(t, dir, "reload")

	cl.Reload()

	newSerial := certLoaderSerial(cl)
	if newSerial == nil {
		t.Fatal("unable to parse certificate serial after reload")
	}
	if origSerial.Cmp(newSerial) == 0 {
		t.Fatal("certificate serial did not change after reload")
	}
}

// TestCertificateLoaderReloadFailureKeepsOldCert verifies that if Reload fails
// the old certificate is preserved.
func TestCertificateLoaderReloadFailureKeepsOldCert(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := generateTestCert(t, dir, "keep-old")

	cl, err := newCertificateLoader(certFile, keyFile)
	if err != nil {
		t.Fatalf("newCertificateLoader failed: %v", err)
	}

	origSerial := certLoaderSerial(cl)
	if origSerial == nil {
		t.Fatal("unable to parse original certificate serial")
	}

	// Corrupt the key file.
	if err := os.WriteFile(keyFile, []byte("not a valid PEM key"), 0600); err != nil {
		t.Fatalf("failed to write corrupted key file: %v", err)
	}

	cl.Reload()

	curSerial := certLoaderSerial(cl)
	if curSerial == nil {
		t.Fatal("unable to parse certificate serial after failed reload")
	}
	if origSerial.Cmp(curSerial) != 0 {
		t.Fatal("certificate changed after a failed reload")
	}
}

// TestCertificateLoaderNilClientHello verifies getCertificate handles a nil
// *tls.ClientHelloInfo argument.
func TestCertificateLoaderNilClientHello(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := generateTestCert(t, dir, "nil-hello")

	cl, err := newCertificateLoader(certFile, keyFile)
	if err != nil {
		t.Fatalf("newCertificateLoader failed: %v", err)
	}

	cert, err := cl.getCertificate(nil)
	if err != nil {
		t.Fatalf("getCertificate(nil) returned error: %v", err)
	}
	if cert == nil {
		t.Fatal("getCertificate(nil) returned nil")
	}
}

// TestCertificateLoaderClientHelloInfo verifies that getCertificate works
// when called with a real *tls.ClientHelloInfo.
func TestCertificateLoaderClientHelloInfo(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := generateTestCert(t, dir, "sni")

	cl, err := newCertificateLoader(certFile, keyFile)
	if err != nil {
		t.Fatalf("newCertificateLoader failed: %v", err)
	}

	hello := &tls.ClientHelloInfo{
		ServerName: "sni.example.com",
	}
	cert, err := cl.getCertificate(hello)
	if err != nil {
		t.Fatalf("getCertificate with ClientHelloInfo failed: %v", err)
	}
	if cert == nil {
		t.Fatal("getCertificate with ClientHelloInfo returned nil")
	}
}
