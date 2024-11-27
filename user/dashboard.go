package user

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"prodata/api"
	"prodata/database/account"
)

func UserNav(ctx *api.Context, userId string) {
	email := account.GetEmailByUuid(userId)
	data := account.GetUser(email)

	hash := sha256.Sum256([]byte(email))
	avatar := hex.EncodeToString(hash[:])

	ctx.Json(map[string]string{
		"email":      email,
		"first_name": data.FirstName,
		"last_name":  data.LastName,
		"avatar":     "https://gravatar.com/avatar/" + avatar,
	})
}

func RecentServices(ctx *api.Context, userId string) {
	services := account.GetServices(userId)

	if len(services) == 0 {
		ctx.WriteHeader(http.StatusOK)
		ctx.Json([]account.Services{})
		return
	}

	err := ctx.Json(services)
	ok := ctx.IfErrNotNull(err)
	if !ok {
		return
	}
}
