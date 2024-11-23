package tx

import (
	"net/http"
)

func PayHook(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	//account.GetUser()
	//
	//request := payment.Request{
	//	TransactionAmount: 1,
	//	PaymentMethodID: "pix",
	//    Payer: &payment.PayerRequest{
	//		Email:
	//	},
	//}

	//bank.Create(request)
}
