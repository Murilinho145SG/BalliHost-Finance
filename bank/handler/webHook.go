package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

func WebHookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var notification PaymentNotification

	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(notification.Data.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	paymentInfo, err := bank.Client().Get(context.Background(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(paymentInfo.Status)

	w.WriteHeader(http.StatusOK)
}
