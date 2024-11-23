package tx

import (
	"context"
	"fmt"
	"net/http"
	"prodata/api"
	"prodata/bank"
	"strconv"
)

type PaymentNotification struct {
	ID       int    `json:"id"`
	LiveMode bool   `json:"live_mode"`
	Type     string `json:"type"`
	Date     string `json:"date"`
	UserID   string `json:"user_id"`
	Version  string `json:"version"`
	Action   string `json:"action"`
	Data     struct {
		ID string `json:"id"`
	} `json:"data"`
}

func WebHookHandler(ctx *api.Context) {
	var notification PaymentNotification

	if err := ctx.Json(&notification); err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(notification.Data.ID)
	if err != nil {
		ctx.Error(err.Error(), http.StatusBadRequest)
		return
	}

	paymentInfo, err := bank.Client().Get(context.Background(), id)
	if err != nil {
		ctx.Error(err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(paymentInfo.Status)

	ctx.WriteHeader(http.StatusOK)
}
