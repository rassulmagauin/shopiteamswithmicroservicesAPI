package middlewares

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type JwtMiddleware struct {
}

func NewJwtMiddleware() *JwtMiddleware {
	return &JwtMiddleware{}
}

var stringToken string = "super secret key"
var api_key string = "1234567"

func (jm *JwtMiddleware) CreateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Second * 3600).Unix(),
	})
	fmt.Println(token)
	return token.SignedString([]byte(stringToken))
}

func (jm *JwtMiddleware) GetJWT(rw http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Access") != "" {
		if r.Header.Get("Access") == api_key {
			token, err := jm.CreateToken()
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}
			rw.Write([]byte(token))
		}
	}
}

func (jm *JwtMiddleware) CheckJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader != "" {
			token, err := jwt.Parse(authorizationHeader, func(token *jwt.Token) (interface{}, error) {
				_, ok := token.Method.(*jwt.SigningMethodHMAC)
				if !ok {
					rw.WriteHeader(http.StatusUnauthorized)
					rw.Write([]byte("Unauthorized1"))
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(stringToken), nil
			})
			if err != nil || !token.Valid {
				rw.WriteHeader(http.StatusUnauthorized)
				rw.Write([]byte("Unauthorized2"))
				return
			}
			next.ServeHTTP(rw, r)
		} else {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte("Unauthorized3"))
		}
	})
}
