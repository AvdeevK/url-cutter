package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/AvdeevK/url-cutter.git/internal/logger"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"os"
	"time"
)

const cookieName = "bearer"
const tokenExp = time.Hour * 3

var secretKey string

func init() {
	secretKey = os.Getenv("SECRET_KEY")
	if secretKey == "" {
		logger.Log.Warn("system variable is non set or empty, use default key")
		secretKey = "supersecretkey"
	}
}

// Взято из примера урока, структура будет из одного поля.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func GenerateUserID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func SetAuthCookie(w http.ResponseWriter, userID string) error {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return errors.New("token signing error")
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    tokenString,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(tokenExp),
	}
	http.SetCookie(w, cookie)

	return nil
}

func GetAuthCookie(r *http.Request) (string, bool, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil || cookie.Value == "" {
		//Комбинация, когда куки нет.
		return "", false, errors.New("token cookie not found")
	}
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	token, err := jwt.ParseWithClaims(cookie.Value, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secretKey), nil
		})

	if err != nil {
		return "", true, errors.New("unexpected signing method")
	}

	if !token.Valid {
		return "", true, errors.New("invalid token")
	}

	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return "", true, errors.New("token has expired")
	}

	//Комбинация токена, который существует, парсинг без ошибок.
	return claims.UserID, true, nil
}
