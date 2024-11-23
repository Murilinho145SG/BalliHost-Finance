package main

import (
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
	api.Post("/transaction/hook", tx.WebHookHandler)
	http.HandleFunc("/account/auth", account.AuthenticateAdmin(func (w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Acesso concedido, Administrador")
	}))
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return
	}
}
