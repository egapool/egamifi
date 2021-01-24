package main

import (
	"fmt"
	"log"
	"time"

	"github.com/egapool/ftx-fr/internal/client"
	"github.com/go-numb/go-ftx/rest/private/wallet"
)

const USD = "USD"

func main() {
	c := client.NewSubClient("shit").Rest

	for {
		balances, err := c.Balances(&wallet.RequestForBalances{})
		if err != nil {
			log.Fatal(err)
		}
		for _, b := range *balances {
			if b.Coin == USD {
				fmt.Println(b.Total, time.Now().Format(time.UnixDate))
			}
		}
		time.Sleep(time.Minute)
	}
}
