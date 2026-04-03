package cookie

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"slices"
)

// See URL: https://www.alexedwards.net/blog/working-with-cookies-in-go
// See URL: https://github.com/gorilla/securecookie
// See URL: https://github.com/syntaqx/cookie

type Signed struct {
	// Generate using the command: "openssl rand -hex 32"
	secret []byte
	name   []byte

	hashFunc func() hash.Hash
	hashSize int
}

// NewSigned creates a new Signed instance with the specified secret and name.
// It decodes the secret from a hexadecimal string and verifies its length.
// The secret should be a SHA-256 key, which can be generated using the command: "openssl rand -hex 32".
// If the secret cannot be decoded or is not of the expected length, the function panics
func NewSigned(secret, name string) *Signed {
	key, err := hex.DecodeString(secret)
	if err != nil {
		panic(fmt.Errorf("unable to decode secret: %w", err))
	}

	s := &Signed{
		secret: key,
		name:   []byte(name),

		hashFunc: sha256.New,
		hashSize: sha256.Size,
	}
	if len(key) != s.hashSize {
		panic(fmt.Errorf("invalid secret length: got %d, expected %d", len(key), sha256.Size))
	}
	return s
}

// Read retrieves the value of the signed cookie from the HTTP request.
// It returns the decoded value of the cookie or an error if the cookie cannot be read or decoded
func (s *Signed) Read(r *http.Request) (string, error) {
	cookie, err := r.Cookie(string(s.name))
	if err != nil {
		return "", fmt.Errorf("unable to read cookie value: %w", err)
	}
	return s.decode(cookie.Value)
}

func (s *Signed) decode(value string) (string, error) {
	signed, err := base64.URLEncoding.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("unable to decode cookie value: %w", err)
	}

	if len(signed) < s.hashSize {
		return "", fmt.Errorf("invalid cookie value length: got %d, expected at least %d", len(signed), s.hashSize)
	}

	b := signed[s.hashSize:]
	if signature := signed[:s.hashSize]; !hmac.Equal(signature, s.createSignature(b)) {
		return "", fmt.Errorf("invalid cookie value")
	}
	return string(b), nil
}

// Write creates a new signed cookie and writes it to the HTTP response.
// The "name" and "value" fields in options will be ignored as they are derived from the Signed instance
func (s *Signed) Write(w http.ResponseWriter, value string, options *http.Cookie) {
	if options == nil {
		options = &http.Cookie{}
	}
	cookie := &http.Cookie{
		Name:     string(s.name),
		Value:    s.encode(value),
		Path:     options.Path,
		Domain:   options.Domain,
		Expires:  options.Expires,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
		SameSite: options.SameSite,
	}
	http.SetCookie(w, cookie)
}

func (s *Signed) encode(value string) string {
	b := []byte(value)
	signed := slices.Concat(s.createSignature(b), b)
	return base64.URLEncoding.EncodeToString(signed)
}

// Delete removes the signed cookie from the HTTP response by setting its MaxAge to -1.
func (s *Signed) Delete(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:  string(s.name),
		Value: "",

		// NOTE: Ensure the cookie is removed
		MaxAge: -1,
	})
}

func (s *Signed) createSignature(value []byte) []byte {
	mac := hmac.New(s.hashFunc, s.secret)
	mac.Write(s.name)
	mac.Write(value)
	return mac.Sum(nil)
}
