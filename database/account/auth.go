package account

import (
	"math/rand"
	"net/http"
	"os"
	"prodata/api"
	"prodata/logs"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MiddleWare para checar se o token é válido para páginas normais
// como áreas de cliente e informações próprias, segurança básica apenas
// para pessoas normais
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

// MiddleWare para checar se o token é valido para acessar páginas de administradores
// ele pega a partir do role dentro do JWT para ver se é válido, se ele não for ele
// retorna acesso não autorizado
func AuthenticateAdmin(next api.ApiFunc) api.ApiFunc {
	return func(ctx *api.Context) {
		tokenString := ctx.Request.Header.Get("Authorization")
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				ctx.Error("Invalid Token", http.StatusUnauthorized)
				ctx.Logger.LogAndSendSystemMessage("Invalid Token, SigningMethod, ADMIN")
				return nil, nil
			}

			return []byte(os.Getenv("JWTKEY")), nil
		})

		if err != nil || !token.Valid {
			ctx.Error("Invalid Token", http.StatusUnauthorized)
			ctx.Logger.LogAndSendSystemMessage(err.Error())
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		uuid := claims["userId"].(string)
		if !UserExist(uuid) {
			ctx.Error("Invalid Token", http.StatusUnauthorized)
			ctx.Logger.LogAndSendSystemMessage("Invalid token, User does not exist ADMIN")
			return
		}

		role := claims["admin"]
		if role == nil || role == "" {
			ctx.Error("Invalid Token", http.StatusUnauthorized)
			ctx.Logger.LogAndSendSystemMessage("Invalid token for access Admin page IP: " + ctx.Request.RemoteAddr)
			return
		}

		if !IsAdmin(uuid) {
			ctx.Error("Invalid Token", http.StatusUnauthorized)
			ctx.Logger.LogAndSendSystemMessage("Invalid token for access Admin page IP: " + ctx.Request.RemoteAddr)
			return
		}

		next(ctx)
	}
}

func MagicGenerator(email string) string {
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

// Ele gera um token próprio que nunca vai ser igual a partir
// do email da pessoa, sendo feito diversos cálculos a partir
// disso criando um token único para cada requisição, mesmo
// sendo o mesmo email, e ele já salva dentro do banco de dados
// e ele salva dentro da DATA a partir do email pegando o UUID
// e o dispositivo usado para fazer a requisição assim dando replace
// no magic link anterior ou até mesmo criando um novo caso não exista
func MagicLinkGenerator(email string) string {
	return MagicLinkMarker(email, MagicGenerator(email))
}
