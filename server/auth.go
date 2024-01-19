package server

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"aidanwoods.dev/go-paseto"
	"libdb.so/ctxt"
)

// SecretKey is a type alias for a PASETO secret key.
type SecretKey = paseto.V4AsymmetricSecretKey

// NewSecretKey returns a new secret key for signing tokens.
// It wraps around PASETO for convenience.
func NewSecretKey() SecretKey {
	return paseto.NewV4AsymmetricSecretKey()
}

// ParseSecretKey parses a secret key from a byte slice.
// It wraps around PASETO for convenience.
func ParseSecretKey(b []byte) (SecretKey, error) {
	return paseto.NewV4AsymmetricSecretKeyFromBytes(b)
}

// TokenExpiry is the duration for which a token is valid.
const TokenExpiry = 30 * 24 * time.Hour

type authenticatedUser struct {
	Username string `json:"u"`
	TeamName string `json:"t"`
}

func (s *Server) setTokenCookie(w http.ResponseWriter, u authenticatedUser) {
	now := time.Now()
	expires := now.Add(TokenExpiry)

	token := paseto.NewToken()
	token.SetIssuedAt(now)
	token.SetNotBefore(now)
	token.SetExpiration(expires)
	token.SetString("u", u.Username)
	token.SetString("t", u.TeamName)

	signed := token.V4Sign(s.secretKey, nil)

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    signed,
		Expires:  expires,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	parser := paseto.NewParser()
	public := s.secretKey.Public()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		token, err := parser.ParseV4Public(public, cookie.Value, nil)
		if err != nil {
			s.logger.Warn(
				"failed to parse token",
				"token", cookie.Value,
				"err", err)
			next.ServeHTTP(w, r)
			return
		}

		var u authenticatedUser
		if err := json.Unmarshal(token.ClaimsJSON(), &u); err != nil {
			s.logger.Warn(
				"failed to parse token",
				"token", cookie.Value,
				"err", err)
			next.ServeHTTP(w, r)
			return
		}

		ctx := ctxt.With(r.Context(), u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	})
}

func isAuthenticated(r *http.Request) bool {
	_, ok := ctxt.From[authenticatedUser](r.Context())
	return ok
}

func getAuthentication(r *http.Request) authenticatedUser {
	v, _ := ctxt.From[authenticatedUser](r.Context())
	return v
}

func generateInviteCode() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	const word = 4
	const count = 4

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	var b strings.Builder
	b.Grow(word*count + count - 1)

	for i := 0; i < count; i++ {
		if i != 0 {
			b.WriteByte('-')
		}
		for j := 0; j < word; j++ {
			b.WriteByte(letters[r.Intn(len(letters))])
		}
	}

	return b.String()
}
