package bank

import (
	"context"

	"github.com/mercadopago/sdk-go/pkg/payment"
)

func Create(request payment.Request) (*payment.Response, error) {
	paymentInfo, err := Client().Create(context.Background(), request)
	if err != nil {
		return paymentInfo, err
	}

	return paymentInfo, nil
}
