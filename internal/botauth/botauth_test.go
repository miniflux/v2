// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package botauth

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestComputeThumbprint(t *testing.T) {
	// Test values taken from https://www.rfc-editor.org/rfc/rfc8037.html#appendix-A.3
	jwk := &jsonWebKey{
		KeyType:   "OKP",
		Curve:     "Ed25519",
		PublicKey: "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo",
	}

	expectedThumbprint := "kPrK_qmxVWaYVA9wwBF6Iuo3vVzz7TxHCTwXBygrS4k"

	thumbprint, err := computeJWKThumbprint(jwk)
	if err != nil {
		t.Fatal(err)
	}

	if thumbprint != expectedThumbprint {
		t.Fatalf("Invalid thumbprint, got %q instead of %q", thumbprint, expectedThumbprint)
	}
}

func TestGenerateSignatureParams(t *testing.T) {
	// Example taken from https://www.rfc-editor.org/rfc/rfc9421#name-signing-a-request-using-ed2
	signatureComponents := []signatureComponent{
		{name: "date", value: "Tue, 20 Apr 2021 02:07:55 GMT"},
		{name: "@method", value: "POST"},
		{name: "@path", value: "/foo"},
		{name: "@authority", value: "example.com"},
		{name: "content-type", value: "application/json"},
		{name: "content-length", value: "18"},
	}

	signatureMetadata := []signatureMetadata{
		{name: "created", value: int64(1618884473)},
		{name: "keyid", value: "test-key-ed25519"},
	}

	generatedSignatureParams := generateSignatureParams(signatureComponents, signatureMetadata)
	expectedSignatureParams := `("date" "@method" "@path" "@authority" "content-type" "content-length");created=1618884473;keyid="test-key-ed25519"`

	if generatedSignatureParams != expectedSignatureParams {
		t.Fatalf("Invalid signature params, got %s instead of %s", generatedSignatureParams, expectedSignatureParams)
	}
}

func TestSignComponents(t *testing.T) {
	// Test key from https://www.rfc-editor.org/rfc/rfc9421#name-example-ed25519-test-key
	privateKeyBase64 := "n4Ni-HpISpVObnQMW0wOhCKROaIKqKtW_2ZYb2p9KcU"
	privateKey, err := base64.RawURLEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		t.Fatal(err)
	}

	// Example taken from https://www.rfc-editor.org/rfc/rfc9421#name-signing-a-request-using-ed2
	signatureComponents := []signatureComponent{
		{name: "date", value: "Tue, 20 Apr 2021 02:07:55 GMT"},
		{name: "@method", value: "POST"},
		{name: "@path", value: "/foo"},
		{name: "@authority", value: "example.com"},
		{name: "content-type", value: "application/json"},
		{name: "content-length", value: "18"},
	}

	signatureMetadata := []signatureMetadata{
		{name: "created", value: int64(1618884473)},
		{name: "keyid", value: "test-key-ed25519"},
	}

	generatedSignatureParams := generateSignatureParams(signatureComponents, signatureMetadata)
	signature, err := signComponents(ed25519.NewKeyFromSeed(privateKey), signatureComponents, generatedSignatureParams)
	if err != nil {
		t.Fatal(err)
	}

	// Expected signature taken from https://www.rfc-editor.org/rfc/rfc9421#name-signing-a-request-using-ed2
	expectedSignature := "wqcAqbmYJ2ji2glfAMaRy4gruYYnx2nEFN2HN6jrnDnQCK1u02Gb04v9EDgwUPiu4A0w6vuQv5lIp5WPpBKRCw=="

	if signature != expectedSignature {
		t.Fatalf("Invalid signature, got %q instead of %q", signature, expectedSignature)
	}
}

func TestServeDirectoryHandler(t *testing.T) {
	// Test keys from https://www.rfc-editor.org/rfc/rfc9421#name-example-ed25519-test-key
	privateKeyBase64 := "n4Ni-HpISpVObnQMW0wOhCKROaIKqKtW_2ZYb2p9KcU"
	privateKeyDecoded, err := base64.RawURLEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := ed25519.NewKeyFromSeed(privateKeyDecoded)

	publicKeyBase64 := "JrQLj5P_89iXES9-vFgrIy29clF9CC_oPPsw3c5D0bs"
	publicKeyDecoded, err := base64.RawURLEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := ed25519.PublicKey(publicKeyDecoded)

	keyPair, err := NewKeyPair(privateKey, publicKey)
	if err != nil {
		t.Fatal(err)
	}

	botAuth, err := NewBothAuth("https://example.com/", KeyPairs{keyPair})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/.well-known/http-message-signatures-directory", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(botAuth.ServeKeyDirectory)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedBody := `{"keys":[{"kty":"OKP","crv":"Ed25519","x":"JrQLj5P_89iXES9-vFgrIy29clF9CC_oPPsw3c5D0bs"}]}`
	if rr.Body.String() != expectedBody {
		t.Fatalf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
	}

	expectedContentType := "application/http-message-signatures-directory+json"
	if rr.Header().Get("Content-Type") != expectedContentType {
		t.Fatalf("handler returned unexpected content type: got %v want %v", rr.Header().Get("Content-Type"), expectedContentType)
	}

	expectedCacheControl := "max-age=86400"
	if rr.Header().Get("Cache-Control") != expectedCacheControl {
		t.Fatalf("handler returned unexpected cache control: got %v want %v", rr.Header().Get("Cache-Control"), expectedCacheControl)
	}

	signatureHeaderValue := rr.Header().Get("Signature")
	if signatureHeaderValue == "" {
		t.Fatal("handler did not return a Signature header")
	}

	if !strings.HasPrefix(signatureHeaderValue, "sig1=:") || !strings.HasSuffix(signatureHeaderValue, ":") {
		t.Fatalf("handler returned unexpected signature: got %v", signatureHeaderValue)
	}

	expectedSignatureInputPrefix := `sig1=("@authority");alg="ed25519";keyid="poqkLGiymh_W0uP6PZFw-dvez3QJT5SolqXBCW38r0U";tag="http-message-signatures-directory";created=`
	signatureInput := rr.Header().Get("Signature-Input")
	if !strings.HasPrefix(signatureInput, expectedSignatureInputPrefix) {
		t.Fatalf("handler returned unexpected signature input: got %v want prefix %v", signatureInput, expectedSignatureInputPrefix)
	}
}

func TestSignRequest(t *testing.T) {
	// Test keys from https://www.rfc-editor.org/rfc/rfc9421#name-example-ed25519-test-key
	privateKeyBase64 := "n4Ni-HpISpVObnQMW0wOhCKROaIKqKtW_2ZYb2p9KcU"
	privateKeyDecoded, err := base64.RawURLEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := ed25519.NewKeyFromSeed(privateKeyDecoded)

	publicKeyBase64 := "JrQLj5P_89iXES9-vFgrIy29clF9CC_oPPsw3c5D0bs"
	publicKeyDecoded, err := base64.RawURLEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := ed25519.PublicKey(publicKeyDecoded)

	keyPair, err := NewKeyPair(privateKey, publicKey)
	if err != nil {
		t.Fatal(err)
	}

	botAuth, err := NewBothAuth("https://signature-agent.test", KeyPairs{keyPair})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "https://example.org", nil)
	if err != nil {
		t.Fatal(err)
	}

	err = botAuth.SignRequest(req)
	if err != nil {
		t.Fatal(err)
	}

	signatureAgentHeaderValue := req.Header.Get("Signature-Agent")
	if signatureAgentHeaderValue != `"https://signature-agent.test"` {
		t.Fatalf("request has unexpected Signature-Agent header: got %v", signatureAgentHeaderValue)
	}

	signatureHeaderValue := req.Header.Get("Signature")
	if signatureHeaderValue == "" {
		t.Fatal("request did not get a Signature header")
	}

	if !strings.HasPrefix(signatureHeaderValue, "sig1=:") || !strings.HasSuffix(signatureHeaderValue, ":") {
		t.Fatalf("request has unexpected signature: got %v", signatureHeaderValue)
	}

	expectedSignatureInputPrefix := `sig1=("@authority" "signature-agent");alg="ed25519";keyid="poqkLGiymh_W0uP6PZFw-dvez3QJT5SolqXBCW38r0U";tag="web-bot-auth";created=`
	signatureInput := req.Header.Get("Signature-Input")
	if !strings.HasPrefix(signatureInput, expectedSignatureInputPrefix) {
		t.Fatalf("request has unexpected signature input: got %v want prefix %v", signatureInput, expectedSignatureInputPrefix)
	}
}
