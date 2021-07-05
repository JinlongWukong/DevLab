package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//jwt services
type JWTService interface {
	GenerateToken(name string, isUser bool, ExpiresAt int64) string
	ValidateToken(token string) (*jwt.Token, error)
}

type CustomClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

type jwtServices struct {
	secretKey string
	issuer    string
}

//auth-jwt
func JWTAuthService() JWTService {
	key := func() string {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "secret"
		}
		return secret
	}()

	return &jwtServices{
		secretKey: key,
		issuer:    "DevLab",
	}
}

func (service *jwtServices) GenerateToken(name string, isUser bool, ExpiresAt int64) string {
	claims := &CustomClaims{
		name,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(ExpiresAt * int64(time.Second))).Unix(),
			Issuer:    service.issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//encoded string
	encodedToken, err := token.SignedString([]byte(service.secretKey))
	if err != nil {
		log.Println(err)
		return ""
	}

	return encodedToken
}

func (service *jwtServices) ValidateToken(encodedToken string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(encodedToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("Invalid token %v", token.Header["alg"])
		}
		return []byte(service.secretKey), nil
	})
}
