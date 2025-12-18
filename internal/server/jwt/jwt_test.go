package jwt

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/msorokin-hash/passkeeper/internal/entity"
	errorscustom "github.com/msorokin-hash/passkeeper/internal/errors"
)

func TestTokenData_CreateToken(t *testing.T) {
	secretKey := "test-secret-key"
	lifetime := 1 * time.Hour
	tokenData := &TokenData{
		lifeTime:  lifetime,
		secretKey: secretKey,
	}

	validUser := &entity.User{
		ID:    "123",
		Login: "testuser",
	}

	t.Run("Success", func(t *testing.T) {
		tokenString, err := tokenData.CreateToken(validUser)
		if err != nil {
			t.Fatalf("CreateToken failed: %v", err)
		}

		if tokenString == "" {
			t.Error("Token string should not be empty")
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil {
			t.Fatalf("Failed to parse token: %v", err)
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			if claims.UserID != validUser.ID {
				t.Errorf("UserID mismatch: expected %s, got %s", validUser.ID, claims.UserID)
			}
			if claims.Login != validUser.Login {
				t.Errorf("Login mismatch: expected %s, got %s", validUser.Login, claims.Login)
			}
		} else {
			t.Error("Invalid token claims")
		}
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		user := &entity.User{
			ID:    "",
			Login: "testuser",
		}

		_, err := tokenData.CreateToken(user)
		fmt.Println(err)
		if err == nil {
			t.Error("Expected error for empty user ID")
		}
		if !errors.Is(err, errorscustom.ErrInvalidUserData) {
			t.Errorf("Expected 'not valid user data' error, got: %v", err)
		}
	})

	t.Run("EmptyLogin", func(t *testing.T) {
		user := &entity.User{
			ID:    "123",
			Login: "",
		}

		_, err := tokenData.CreateToken(user)
		if err == nil {
			t.Error("Expected error for empty login")
		}
	})
}

func TestTokenData_GetClaims(t *testing.T) {
	secretKey := "test-secret-key"
	lifetime := 1 * time.Hour
	tokenData := &TokenData{
		lifeTime:  lifetime,
		secretKey: secretKey,
	}

	validUser := &entity.User{
		ID:    "123",
		Login: "testuser",
	}

	validToken, err := tokenData.CreateToken(validUser)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	t.Run("Success", func(t *testing.T) {
		claims, err := tokenData.GetClaims(validToken)
		if err != nil {
			t.Fatalf("GetClaims failed: %v", err)
		}

		if claims.UserID != validUser.ID {
			t.Errorf("UserID mismatch: expected %s, got %s", validUser.ID, claims.UserID)
		}
		if claims.Login != validUser.Login {
			t.Errorf("Login mismatch: expected %s, got %s", validUser.Login, claims.Login)
		}
	})

	t.Run("InvalidSecretKey", func(t *testing.T) {
		wrongTokenData := &TokenData{
			lifeTime:  lifetime,
			secretKey: "wrong-secret-key",
		}

		_, err := wrongTokenData.GetClaims(validToken)
		if err == nil {
			t.Error("Expected error for invalid secret key")
		}
	})

	t.Run("MalformedToken", func(t *testing.T) {
		malformedToken := validToken + "corrupt"
		_, err := tokenData.GetClaims(malformedToken)
		if err == nil {
			t.Error("Expected error for malformed token")
		}
	})

	t.Run("EmptyToken", func(t *testing.T) {
		_, err := tokenData.GetClaims("")
		if err == nil {
			t.Error("Expected error for empty token")
		}
	})

	t.Run("WrongSigningMethod", func(t *testing.T) {
		rsaToken := jwt.NewWithClaims(jwt.SigningMethodRS256, Claims{
			UserID: "123",
			Login:  "testuser",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(lifetime)),
			},
		})

		rsaTokenString, _ := rsaToken.SignedString([]byte("random-key"))

		_, err := tokenData.GetClaims(rsaTokenString)
		if err == nil {
			t.Error("Expected error for wrong signing method")
		}
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		expiredTokenData := &TokenData{
			lifeTime:  -1 * time.Hour,
			secretKey: secretKey,
		}

		expiredToken, err := expiredTokenData.CreateToken(validUser)
		if err != nil {
			t.Fatalf("Failed to create expired token: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		_, err = tokenData.GetClaims(expiredToken)
		if err == nil {
			t.Error("Expected error for expired token")
		}
	})
}
