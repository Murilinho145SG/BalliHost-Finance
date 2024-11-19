package main

import (
	"fmt"
	"net/http"
	"prodata/bank/handler"
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
	
	http.HandleFunc("/account/login", user.HandlerLogin)
	http.HandleFunc("/account/verify/", user.HandlerMagicLink)
	http.HandleFunc("/account/register", user.HandlerRegister)
	http.HandleFunc("/account/auth", account.AuthenticateAdmin(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Acesso concedido, Administrador")
	}))
	http.HandleFunc("/mercadopago/webhook", handler.WebHookHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		logs.NewSistemLogger().LogAndSendSystemMessage(err.Error())
		return
	}
}
