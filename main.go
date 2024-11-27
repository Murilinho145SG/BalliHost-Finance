package main

import (
	"fmt"
	"net/http"
	"prodata/api"
	"prodata/bank/tx"
	"prodata/database/account"
	"prodata/logs"
	"prodata/user"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	account.AddServices("01fd92c3-a8cc-4664-9851-4914a5b49842", &account.Services{
		Id:     uuid.New().String(),
		Name:   "Minecraft Premium 48GB",
		Price:  float64(480),
		Status: "Ativo",
		Date:   time.Now().Add(30 * time.Hour * 24).Format(time.DateTime),
		Type:   "Hospedagem de Jogos",
	})

	api.Post("/account/login", user.HandlerLogin)
	api.Post("/account/register", user.HandlerRegister)
	api.Post("/account/verify/", user.HandlerMagicLink)
	api.Post("/account/auth/generate", user.HandlerNewMagicLink)
	api.Post("/account/query-password", user.HandlerMakePasswordResetPage)
	api.Post("/account/reset-password/", user.HandlerChangePasswordReset)
	api.Get("/dashboard/navbar", account.Authenticate(user.UserNav))
	api.Get("/dashboard/recent-services", account.Authenticate(user.RecentServices))
	api.Post("/information/error", user.HandlerErrors)
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
