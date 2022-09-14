package session

import (
	"errors"
	"os"
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

	signedAccessToken, err := accessToken.SignedString(os.Getenv("JWT_ACCESS_TOKEN_SECRET_KEY"))

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

	signedRefreshToken, err := refreshToken.SignedString(os.Getenv("JWT_REFRESH_TOKEN_SECRET_KEY"))

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
	atToken, err := jwt.ParseWithClaims(signedAccessToken, atClaims, func(t *jwt.Token) (interface{}, error) {
		return os.Getenv("JWT_ACCESS_TOKEN_SECRET_KEY"), nil
	})

	if err != nil {
		return nil, err
	}

	rtClaims := &RefreshTokenClaims{}
	rtToken, err := jwt.ParseWithClaims(signedRefreshToken, rtClaims, func(t *jwt.Token) (interface{}, error) {
		return os.Getenv("JWT_REFRESH_TOKEN_SECRET_KEY"), nil
	})

	if err != nil {
		return nil, err
	}

	if !atToken.Valid || !rtToken.Valid {
		return nil, errors.New("Invalid token sent")
	}

	if atClaims.AccountName != rtClaims.AccountName || atClaims.EmailId != rtClaims.Emailid {
		return nil, errors.New("Invalid token sent")
	}

	if time.Unix(atClaims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		return nil, errors.New("Too early to renew")
	}

	newAtExpirationTime := time.Now().Add(15 * time.Minute)
	atClaims.StandardClaims.ExpiresAt = newAtExpirationTime.Unix()
	
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	newSignedAccessToken, err := accessToken.SignedString(os.Getenv("JWT_ACCESS_TOKEN_SECRET_KEY"))

	return &JwtTokens{AccessTokenSigned: newSignedAccessToken}, nil
}

func VerifyAccessToken(signedAccessToken string) (*Payload, bool) {
	atClaims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(signedAccessToken, atClaims, func(t *jwt.Token) (interface{}, error) {
		return os.Getenv("JWT_ACCESS_TOKEN_SECRET_KEY"), nil
	})

	if err != nil {
		return nil, false
	}

	if !token.Valid {
		return nil, false
	}
	
	return &Payload{
		AccountName: atClaims.AccountName,
		EmailId:     atClaims.EmailId,
	}, false
}
