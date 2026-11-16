package examples

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
)

// maxState is the largest cookie value a demo may keep. A visitor who fills it
// is answered with a refusal by the demo; nothing grows.
const maxState = 1800

// store keeps a demo's state in a signed cookie of the visitor.
type store struct {
	key    []byte
	secure bool
}

func newStore(key []byte, secure bool) store {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("ypy.dev/examples"))
	return store{key: mac.Sum(nil), secure: secure}
}

func cookieName(name string) string { return "ypy_demo_" + strings.ReplaceAll(name, "-", "_") }

func (s store) sign(name, payload string) string {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(name + "." + payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// read fills v from the visitor's cookie and says whether there was a valid one.
func (s store) read(r *http.Request, name string, v any) bool {
	c, err := r.Cookie(cookieName(name))
	if err != nil || len(c.Value) > maxState+64 {
		return false
	}
	payload, sig, ok := strings.Cut(c.Value, ".")
	if !ok || !hmac.Equal([]byte(sig), []byte(s.sign(name, payload))) {
		return false
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	return err == nil && json.Unmarshal(raw, v) == nil
}

// write puts v in the cookie. It refuses, and says so, when v is too large.
func (s store) write(w http.ResponseWriter, name string, v any) bool {
	raw, err := json.Marshal(v)
	if err != nil {
		return false
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	if len(payload) > maxState {
		return false
	}
	http.SetCookie(w, &http.Cookie{
		Name: cookieName(name), Value: payload + "." + s.sign(name, payload),
		Path: "/", MaxAge: 3600, HttpOnly: true, Secure: s.secure, SameSite: http.SameSiteLaxMode,
	})
	return true
}
