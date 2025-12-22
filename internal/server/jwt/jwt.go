package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/msorokin-hash/passkeeper/internal/entity"
	errorscustom "github.com/msorokin-hash/passkeeper/internal/errors"
)

//go:generate mockgen -source=jwt.go -destination=mocks/token_mock.go -package=mocks
type TokenService interface {
	CreateToken(user *entity.User) (string, error)
	GetClaims(tokenString string) (*Claims, error)
}

// Claims defines the payload stored inside a JWT token. It embeds
// jwt.RegisteredClaims and adds application-specific fields such as
// the user ID and login.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
	Login  string
}

// TokenData represents the configuration and behavior of the token
// generation service. It stores the token lifetime and the secret key
// used for signing and validating JWT tokens.
type TokenData struct {
	lifeTime  time.Duration
	secretKey string
}

// NewTokenDataService creates and returns a new TokenData instance.
// The provided tokenKey is used for signing JWTs, and lifeTime defines
// the token validity duration in minutes.
func NewTokenDataService(tokenKey string, lifeTime time.Duration) *TokenData {
	return &TokenData{
		lifeTime:  lifeTime * time.Minute,
		secretKey: tokenKey,
	}
}

// CreateToken generates a signed JWT token for the given user.
// It embeds the user's ID and login into the token claims and sets
// an expiration time based on the configured token lifetime.
// Returns ErrInvalidUserData if required user fields are empty.
// Any signing error is wrapped and returned.
func (t *TokenData) CreateToken(user *entity.User) (string, error) {
	if user.ID == "" || user.Login == "" {
		return "", errorscustom.ErrInvalidUserData
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: user.ID,
		Login:  user.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.lifeTime)),
		},
	})

	tokenString, err := token.SignedString([]byte(t.secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to build jwt string: %w", err)
	}

	return tokenString, nil
}

// GetClaims parses the provided JWT string and returns the extracted
// Claims structure. It validates the token signature and ensures that
// the token was signed using an HMAC algorithm. If the token is invalid,
// Expired, or cannot be parsed, the method returns an error such as
// ErrInvalidJwtToken or a wrapped parsing error.
func (t *TokenData) GetClaims(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse jwt token: %w", err)
	}

	if !token.Valid {
		return nil, errorscustom.ErrInvalidJwtToken
	}

	return claims, nil
}
