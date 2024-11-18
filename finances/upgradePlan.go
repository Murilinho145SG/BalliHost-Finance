package finances

import (
	"errors"
	"fmt"
	"math"
	"time"
)

func GeneratePrice(date time.Time, oldPrice float64, newPrice float64) (float64, error) {
	now := time.Now()

	if date.After(now) {
		return 0, errors.New("is not valid date")
	}

	days := now.Sub(date).Hours() / 24
	fmt.Println(days)

	if days == 0 {
		return newPrice, nil
	}

	if date.Month() != now.Month() && date.Day() == now.Day() {
		return newPrice, nil
	}

	oldPriceInDays := oldPrice / 30

	consumedPrice := oldPriceInDays * days

	price := newPrice - consumedPrice

	price = math.Round(price*100) / 100

	return price, nil
}
