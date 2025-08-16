package services

import (
	"SalaryAdvance/internal/domain"
	"SalaryAdvance/pkg/config"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTServiceImpl struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTService(cfg *config.Config) domain.JWTService {
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey
	var err error

	
	if cfg.PrivateKeyPEM != "" && cfg.PublicKeyPEM != "" {
		privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(cfg.PrivateKeyPEM))
		if err != nil {
			fmt.Printf("Failed to parse private key: %v\n", err)
			panic(err)
		}
		publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(cfg.PublicKeyPEM))
		if err != nil {
			fmt.Printf("Failed to parse public key: %v\n", err)
			panic(err)
		}
	} else {
		
		privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			fmt.Printf("Failed to generate private key: %v\n", err)
			panic(err)
		}
		publicKey = &privateKey.PublicKey
		fmt.Println("Generated new RSA key pair (development mode)")
	}

	return &JWTServiceImpl{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     cfg.Issuer,
		accessTTL:  time.Duration(cfg.AccessTTLMin) * time.Minute,
		refreshTTL: time.Duration(cfg.RefreshTTLMin) * time.Minute,
	}
}

func (s *JWTServiceImpl) GenerateAccessToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
		"iss":   s.issuer,
		"exp":   time.Now().Add(s.accessTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		fmt.Printf("Failed to sign access token: %v\n", err)
		return "", err
	}
	return signed, nil
}

func (s *JWTServiceImpl) GenerateRefreshToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"id":  user.ID,
		"iss": s.issuer,
		"exp": time.Now().Add(s.refreshTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		fmt.Printf("Failed to sign refresh token: %v\n", err)
		return "", err
	}
	return signed, nil
}

func (s *JWTServiceImpl) GenerateInviteToken(email string, adminID uint) (string, error) {
	claims := jwt.MapClaims{
		"email":      email,
		"invited_by": adminID,
		"iss":        s.issuer,
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(s.privateKey)
	if err != nil {
		fmt.Printf("Failed to sign invite token: %v\n", err)
		return "", err
	}
	return signed, nil
}

func (s *JWTServiceImpl) ValidateToken(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			fmt.Printf("Unexpected signing method: %v\n", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})
	if err != nil {
		fmt.Printf("Token parsing failed: %v\n", err)
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		fmt.Println("Invalid token claims or token not valid")
		return nil, fmt.Errorf("invalid token")
	}

	if claims["iss"] != s.issuer {
		fmt.Printf("Invalid issuer: got %v, expected %v\n", claims["iss"], s.issuer)
		return nil, fmt.Errorf("invalid issuer")
	}
	return claims, nil
}
