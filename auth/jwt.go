package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/SevenTV/Common/utils"
	"github.com/golang-jwt/jwt/v4"
)

func SignJWT(secret string, claim jwt.Claims) (string, error) {
	// Generate an unsigned token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	// Sign the token
	tokenStr, err := token.SignedString(utils.S2B(secret))

	return tokenStr, err
}

type JWTClaimUser struct {
	UserID       string  `json:"u"`
	TokenVersion float64 `json:"v"`

	jwt.RegisteredClaims
}

type JWTClaimOAuth2CSRF struct {
	State     string    `json:"s"`
	CreatedAt time.Time `json:"at"`

	jwt.RegisteredClaims
}

func VerifyJWT(secret string, token []string) (*jwt.Token, jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	result, err := jwt.ParseWithClaims(
		strings.Join(token, "."),
		claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("bad jwt signing method, expected HMAC but got %v", t.Header["alg"])
			}

			return utils.S2B(secret), nil
		},
	)

	return result, claims, err
}
