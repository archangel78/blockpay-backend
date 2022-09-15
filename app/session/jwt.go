package session

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

type Payload struct {
	AccountName string `json:"accountName"`
	EmailId     string `json:"emailId"`
}

type AccessTokenClaims struct {
	AccountName string `json:"accountName"`
	EmailId     string `json:"emailId"`
	jwt.StandardClaims
}

type RefreshTokenClaims struct {
	AccountName string `json:"accountName"`
	Emailid     string `josn:"emailId"`
	Refreshes   int    `json:"Refreshes"`
	jwt.StandardClaims
}

type JwtTokens struct {
	AccessTokenSigned  string
	RefreshTokenSigned string
}

func GenerateTokenPair(accountName string, emailId string) (*JwtTokens, error) {
	// Create access token
	atExpirationTime := time.Now().Add(15 * time.Minute)

	accessTokenClaims := &AccessTokenClaims{
		AccountName: accountName,
		EmailId:     emailId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: atExpirationTime.Unix(),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	signedAccessToken, err := accessToken.SignedString([]byte(os.Getenv("JWT_ACCESS_TOKEN_SECRET_KEY")))

	if err != nil {
		return nil, err
	}

	// Create refresh token
	refExpirationTime := time.Now().Add(720 * time.Hour)

	refreshTokenClaims := &RefreshTokenClaims{
		AccountName: accountName,
		Emailid:     emailId,
		Refreshes:   0,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: refExpirationTime.Unix(),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	signedRefreshToken, err := refreshToken.SignedString([]byte(os.Getenv("JWT_REFRESH_TOKEN_SECRET_KEY")))

	if err != nil {
		return nil, err
	}

	return &JwtTokens{
		AccessTokenSigned:  signedAccessToken,
		RefreshTokenSigned: signedRefreshToken,
	}, nil
}

func RenewAccessToken(signedAccessToken string, signedRefreshToken string) (*JwtTokens, error) {
	atClaims := &AccessTokenClaims{}
	_, err := jwt.ParseWithClaims(signedAccessToken, atClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_ACCESS_TOKEN_SECRET_KEY")), nil
	})

	if err != nil {
		if !strings.Contains(err.Error(), "expired"){
			return nil, err
		}
	}

	rtClaims := &RefreshTokenClaims{}
	rtToken, err := jwt.ParseWithClaims(signedRefreshToken, rtClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_REFRESH_TOKEN_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, err
	}

	if !rtToken.Valid {
		return nil, errors.New("Invalid refresh token")
	}

	if atClaims.AccountName != rtClaims.AccountName || atClaims.EmailId != rtClaims.Emailid {
		return nil, errors.New("Different claims in refresh token and access token")
	}

	if time.Unix(atClaims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		return nil, errors.New("Too early to renew")
	}

	newAtExpirationTime := time.Now().Add(15 * time.Minute)
	atClaims.StandardClaims.ExpiresAt = newAtExpirationTime.Unix()
	
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	newSignedAccessToken, err := accessToken.SignedString([]byte(os.Getenv("JWT_ACCESS_TOKEN_SECRET_KEY")))

	return &JwtTokens{AccessTokenSigned: newSignedAccessToken}, nil
}

func VerifyAccessToken(signedAccessToken string) (*Payload, bool, error) {
	atClaims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(signedAccessToken, atClaims, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_ACCESS_TOKEN_SECRET_KEY")), nil
	})

	if err != nil {
		return nil, false, err
	}

	if !token.Valid {
		return nil, false, errors.New("Invalid token")
	}
	
	return &Payload{
		AccountName: atClaims.AccountName,
		EmailId:     atClaims.EmailId,
	}, true, nil
}
