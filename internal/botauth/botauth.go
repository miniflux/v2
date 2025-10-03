// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package botauth // import "miniflux.app/v2/internal/botauth"

// Resources:
//
// https://datatracker.ietf.org/doc/html/draft-meunier-http-message-signatures-directory
// https://datatracker.ietf.org/doc/html/draft-meunier-web-bot-auth-architecture
// https://developers.cloudflare.com/bots/reference/bot-verification/web-bot-auth/
// https://github.com/thibmeu/http-message-signatures-directory

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"miniflux.app/v2/internal/crypto"
)

const (
	signatureValidity = 3600 // 1 hour validity
)

var GlobalInstance *botAuth

type jsonWebKey struct {
	KeyType   string `json:"kty"`
	Curve     string `json:"crv"`
	PublicKey string `json:"x"`
}

type jsonWebKeySet struct {
	Keys []jsonWebKey `json:"keys"`
}

type keyPair struct {
	privateKey []byte
	publicKey  []byte
	publicJWK  *jsonWebKey
	thumbprint string
}

func NewKeyPair(privateKey, publicKey []byte) (*keyPair, error) {
	if len(privateKey) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: got %d instead of %d", len(privateKey), ed25519.PrivateKeySize)
	}
	if len(publicKey) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: got %d instead of %d", len(publicKey), ed25519.PublicKeySize)
	}

	publicJWK := &jsonWebKey{
		KeyType:   "OKP",
		Curve:     "Ed25519",
		PublicKey: base64.RawURLEncoding.EncodeToString(publicKey),
	}

	thumbprint, err := computeJWKThumbprint(publicJWK)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate JWK thumbprint: %w", err)
	}

	return &keyPair{
		privateKey: privateKey,
		publicKey:  publicKey,
		publicJWK:  publicJWK,
		thumbprint: thumbprint,
	}, nil
}

type KeyPairs []*keyPair

func (kps KeyPairs) jsonWebKeySet() jsonWebKeySet {
	var keys []jsonWebKey
	for _, kp := range kps {
		keys = append(keys, *kp.publicJWK)
	}
	return jsonWebKeySet{Keys: keys}
}

type botAuth struct {
	directoryURL string
	keys         KeyPairs
}

func NewBothAuth(directoryURL string, keys KeyPairs) (*botAuth, error) {
	if !strings.HasPrefix(directoryURL, "https://") {
		return nil, fmt.Errorf("directory URL %q must start with https://", directoryURL)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("at least one key pair is required")
	}

	return &botAuth{
		directoryURL: directoryURL,
		keys:         keys,
	}, nil
}

func (ba *botAuth) DirectoryURL() string {
	absoluteURL, err := url.JoinPath(ba.directoryURL, "/.well-known/http-message-signatures-directory")
	if err != nil {
		return ba.directoryURL
	}
	return absoluteURL
}

func (ba *botAuth) ServeKeyDirectory(w http.ResponseWriter, r *http.Request) {
	body, err := json.Marshal(ba.keys.jsonWebKeySet())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	created := time.Now().Unix()
	expires := created + signatureValidity
	signatures := make([]string, len(ba.keys))
	signatureInputs := make([]string, len(ba.keys))

	for i, key := range ba.keys {
		signatureMetadata := []signatureMetadata{
			{name: "alg", value: "ed25519"},

			// https://datatracker.ietf.org/doc/html/draft-meunier-http-message-signatures-directory-01#section-5.2-6.6.1
			{name: "keyid", value: key.thumbprint},

			// https://datatracker.ietf.org/doc/html/draft-meunier-http-message-signatures-directory-01#section-5.2-6.8.1
			{name: "tag", value: "http-message-signatures-directory"},

			// https://datatracker.ietf.org/doc/html/draft-meunier-http-message-signatures-directory-01#section-5.2-6.2.1
			{name: "created", value: created},

			// https://datatracker.ietf.org/doc/html/draft-meunier-http-message-signatures-directory-01#section-5.2-6.4.1
			{name: "expires", value: expires},
		}

		signatureComponents := []signatureComponent{
			{name: "@authority", value: r.Host},
		}

		signatureParams := generateSignatureParams(signatureComponents, signatureMetadata)

		signature, err := signComponents(key.privateKey, signatureComponents, signatureParams)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		signatureLabel := `sig` + strconv.Itoa(i+1)
		signatureInputs[i] = signatureLabel + `=` + signatureParams
		signatures[i] = signatureLabel + `=:` + signature + `:`
	}

	// https://datatracker.ietf.org/doc/html/draft-meunier-http-message-signatures-directory-01#name-application-http-message-si
	w.Header().Set("Content-Type", "application/http-message-signatures-directory+json")
	w.Header().Set("Signature-Input", strings.Join(signatureInputs, ", "))
	w.Header().Set("Signature", strings.Join(signatures, ", "))

	// Verifiers can cache keys directory for 1 day.
	w.Header().Set("Cache-Control", "max-age=86400")

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func (ba *botAuth) SignRequest(req *http.Request) error {
	if len(ba.keys) == 0 {
		return fmt.Errorf("no key pairs available to sign the request")
	}

	firstKeyPair := ba.keys[0]
	created := time.Now().Unix()
	expires := created + signatureValidity

	// @authority component
	// https://www.rfc-editor.org/rfc/rfc9421#section-2.2.3
	authority := req.Host
	if authority == "" {
		authority = req.URL.Host
	}

	signatureAgent := `"` + ba.directoryURL + `"`

	signatureMetadata := []signatureMetadata{
		{name: "alg", value: "ed25519"},

		// https://datatracker.ietf.org/doc/html/draft-meunier-web-bot-auth-architecture-02#section-4.2-5.6.1
		{name: "keyid", value: firstKeyPair.thumbprint},

		// https://datatracker.ietf.org/doc/html/draft-meunier-web-bot-auth-architecture-02#section-4.2-5.8.1
		{name: "tag", value: "web-bot-auth"},

		// https://datatracker.ietf.org/doc/html/draft-meunier-web-bot-auth-architecture-02#section-4.2-5.2.1
		{name: "created", value: created},

		// https://datatracker.ietf.org/doc/html/draft-meunier-web-bot-auth-architecture-02#section-4.2-5.4.1
		{name: "expires", value: expires},

		// https://datatracker.ietf.org/doc/html/draft-meunier-web-bot-auth-architecture-02#name-anti-replay
		{name: "nonce", value: base64.StdEncoding.EncodeToString(crypto.GenerateRandomBytes(64))},
	}

	// https://datatracker.ietf.org/doc/html/draft-meunier-web-bot-auth-architecture-02#name-signature-agent
	signatureComponents := []signatureComponent{
		{name: "@authority", value: authority},
		{name: "signature-agent", value: signatureAgent},
	}

	signatureParams := generateSignatureParams(signatureComponents, signatureMetadata)
	signatureInput := `sig1=` + signatureParams

	signature, err := signComponents(firstKeyPair.privateKey, signatureComponents, signatureParams)
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	// Add headers to request
	req.Header.Set("Signature-Agent", signatureAgent)
	req.Header.Set("Signature-Input", signatureInput)
	req.Header.Set("Signature", `sig1=:`+signature+`:`)

	return nil
}

// https://www.rfc-editor.org/rfc/rfc8037.html#appendix-A.3
func computeJWKThumbprint(jwk *jsonWebKey) (string, error) {
	canonical := `{"crv":"` + jwk.Curve + `","kty":"` + jwk.KeyType + `","x":"` + jwk.PublicKey + `"}`
	hash := sha256.Sum256([]byte(canonical))
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

type signatureMetadata struct {
	name  string
	value any
}

// https://www.rfc-editor.org/rfc/rfc9421#name-signature-parameters
func generateSignatureParams(components []signatureComponent, signatureMetadata []signatureMetadata) string {
	var componentNames []string

	for _, component := range components {
		componentNames = append(componentNames, `"`+component.name+`"`)
	}

	var metadataParts []string
	for _, meta := range signatureMetadata {
		switch v := meta.value.(type) {
		case string:
			metadataParts = append(metadataParts, meta.name+`="`+v+`"`)
		case int64:
			metadataParts = append(metadataParts, meta.name+`=`+strconv.FormatInt(v, 10))
		}
	}

	return `(` + strings.Join(componentNames, ` `) + `);` + strings.Join(metadataParts, ";")
}

type signatureComponent struct {
	name  string
	value string
}

// https://www.rfc-editor.org/rfc/rfc9421#name-signing-request-components-
func signComponents(privateKey ed25519.PrivateKey, components []signatureComponent, signatureParams string) (string, error) {
	var signatureBase strings.Builder

	// Build signature base
	for _, comp := range components {
		signatureBase.WriteString(`"` + comp.name + `": ` + comp.value + "\n")
	}

	signatureBase.WriteString(`"@signature-params": ` + signatureParams)

	// Sign the signature base
	signature := ed25519.Sign(privateKey, []byte(signatureBase.String()))

	return base64.StdEncoding.EncodeToString(signature), nil
}
