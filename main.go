package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"prodata/api"
	"prodata/bank/tx"
	"prodata/database/account"
	"prodata/logs"
	"prodata/user"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	api.Post("/account/login", user.HandlerLogin)
	api.Post("/account/register", user.HandlerRegister)
	api.Post("/account/verify/", user.HandlerMagicLink)
	api.Post("/account/auth/generate", user.HandlerNewMagicLink)
	api.Post("/account/query-password", user.HandlerMakePasswordResetPage)
	api.Post("/account/reset-password/", user.HandlerChangePasswordReset)
	api.Get("/dashboard/navbar", account.Authenticate(func(ctx *api.Context, userId string) {
		email := account.GetEmailByUuid(userId)
		data := account.GetUser(email)
		
		hash := sha256.Sum256([]byte(email))
		avatar := hex.EncodeToString(hash[:])

		ctx.Json(map[string]string{
			"email":      email,
			"first_name": data.FirstName,
			"last_name":  data.LastName,
			"avatar": "https://gravatar.com/avatar/" + avatar,
		})
	}))
	api.Post("/transaction/hook", tx.WebHookHandler)
	api.Post("/account/auth", account.AuthenticateAdmin(func(ctx *api.Context) {
		fmt.Fprintln(ctx.Writer, "Acesso concedido, Administrador")
	}))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return
	}
}
