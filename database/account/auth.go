package account

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"prodata/logs"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logs.NewSistemLogger()

		tokenString := r.Header.Get("Authorization")

		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				logger.LogAndSendSystemMessage("Invalid token, SigningMethod")
				return nil, nil
			}
			return []byte(os.Getenv("JWTKEY")), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			logger.LogAndSendSystemMessage(err.Error())
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		uuid := claims["userId"].(string)
		if !UserExist(uuid) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			logger.LogAndSendSystemMessage("Invalid token, User does not exist")
			return
		}

		next(w, r)
	}
}

func ProtectedRoute(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintln(w, "acesso concedido!, bem-vindo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func MagicLinkGenerator(device, email string) string {
	token := func() string {
		allowedChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@$&"

		UEmail := strings.ToUpper(email)

		letters := strings.Split(UEmail, "")
		key := os.Getenv("MAGIC_KEY")

		var token strings.Builder

		now := time.Now()
		day := now.Day()
		sec := now.Second()

		keyT := day * sec
		if keyT > 30 {
			keyT /= 10
		}

		for index, letter := range letters {
			for i := 'A'; i <= 'Z'; i++ {
				if letter == string(i) {
					randomChar := allowedChars[(index+int(keyT))%len(allowedChars)]
					token.WriteByte(byte(randomChar))
				}
			}
		}

		for token.Len() < 30 {
			token.WriteByte(byte(allowedChars[rand.Intn(len(allowedChars))]))
		}

		tokenR := strings.Builder{}
		for index, letter := range token.String() {
			if letter == ' ' || letter == '/' || letter == '\\' || letter == '}' || letter == '{' || letter == '|' {
				letter = rune((int32(index) + int32(key[index%len(key)])*int32(keyT)) / 2 % 127)
			}
			if letter == '/' {
				letter = '_'
			}
			if letter == '\\' {
				letter = '%'
			}

			tokenR.WriteRune(letter)
		}
		
		return strings.ReplaceAll(tokenR.String(), "\x00", "&")
	}

	return MagicLinkMarker(email, device, token())
}