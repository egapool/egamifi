package main

import (
	"fmt"
	"log"
	"time"

	"github.com/egapool/egamifi/internal/client"
	"github.com/go-numb/go-ftx/rest/private/wallet"
)

const USD = "USD"

func main() {
	c := client.NewSubRestClient("ap")
	// c := client.NewRestClient()

	// for {
	balances, err := c.Balances(&wallet.RequestForBalances{})
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range *balances {
		fmt.Println(b, time.Now().Format(time.UnixDate))
		// if b.Coin == USD {
		// }
	}
	// time.Sleep(time.Minute)
	// }
}
