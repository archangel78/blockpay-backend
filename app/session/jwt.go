package session

import (
	"github.com/golang-jwt/jwt"
	"os"
	"time"
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
	atExpirationTime := time.Now().Add(5 * time.Minute)

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

func VerifyAccessToken(signedAccessToken string, accountName string) (*Payload, bool) {
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

	if accountName != atClaims.AccountName {
		return nil, false
	}
	
	return &Payload{
		AccountName: atClaims.AccountName,
		EmailId:     atClaims.EmailId,
	}, false
}
