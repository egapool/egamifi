package bborder

import (
	"fmt"
	"log"

	"github.com/egapool/egamifi/internal/indicators"
	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/private/orders"
	"github.com/go-numb/go-ftx/rest/public/markets"
)

type BbOrder struct {
	client *rest.Client
}

func NewBbOrder(client *rest.Client) *BbOrder {
	return &BbOrder{client: client}
}

func (c *BbOrder) Run(market string) {
	_, err := c.client.CancelAll(&orders.RequestForCancelAll{
		ProductCode: market})
	if err != nil {
		log.Fatal(err)
	}
	var mf indicators.Mfloat

	req := &markets.RequestForCandles{
		ProductCode: market,
		Resolution:  60,
		Limit:       40,
	}
	candles, err := c.client.Candles(req)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range *candles {
		mf = append(mf, c.Close)
	}
	_, upper, lower := indicators.BollingerBands(mf, 20, 2)
	upper_price := upper[len(upper)-1:][0]
	lower_price := lower[len(lower)-1:][0]
	// fmt.Println(upper[len(upper)-1:][0], middle[len(middle)-1:], lower[len(middle)-1:])
	c.PlaceOrder(market, "sell", upper_price, 0.001, 1)
	c.PlaceOrder(market, "buy", lower_price, 0.001, 1)
}

func (c *BbOrder) PlaceOrder(market string, side string, price float64, lot float64, quantity int) {
	base_price := price
	for i := 0; i < quantity; i++ {
		req := &orders.RequestForPlaceOrder{
			Market: market,
			Type:   "limit",
			Side:   side,
			Price:  base_price,
			Size:   lot,
			Ioc:    false}
		res, err := c.client.PlaceOrder(req)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)
		if side == "sell" {
			base_price = base_price * 1.0008
		} else {
			base_price = base_price - (base_price * 0.0008)
		}
	}
}
