package auth

import (
	"fmt"
	"log"

	"github.com/JinlongWukong/DevLab/notification"
	"github.com/JinlongWukong/DevLab/utils"
)

var expiresIn int64

func init() {
	expiresIn = 86400
}

func OneTimePassGen(target string) string {
	password := utils.RandomString(8)
	msg := fmt.Sprintf("From DevLab, your one-time password is:  %v", password)
	notification.SendNotification(notification.Message{Target: target, Text: msg})

	return password
}

func InvokeToken(name string) TokenInfo {
	author := JWTAuthService()

	tokenInfo := TokenInfo{
		AccessToken:       author.GenerateToken(name, true, expiresIn),
		AuthorizationType: "bearer",
		ExpiresIn:         expiresIn,
	}

	return tokenInfo
}

func ValidateToken(tokenString string) (error, string) {
	author := JWTAuthService()

	token, err := author.ValidateToken(tokenString)
	if token != nil && token.Valid {
		claims := token.Claims.(*CustomClaims)
		log.Println(claims)
		return nil, claims.Name
	} else {
		log.Println(err)
		return err, ""
	}
}
