package bank

import (
	"os"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
)

func Client() payment.Client {
	accessToken := os.Getenv("MP_ACCESS_TOKEN")
	cfg, err := config.New(accessToken)
	if err != nil {
		panic(err)
	}

	client := payment.NewClient(cfg)
	return client
}
