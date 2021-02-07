package account

import (
	"fmt"
	"log"

	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/private/wallet"
)

func BalancesAll(c *rest.Client) {
	res, err := c.BalancesAll(&wallet.RequestForBalancesAll{})
	if err != nil {
		log.Fatal(err)
	}
	for account, w := range *res {
		fmt.Println(account)
		for _, balance := range w {
			fmt.Println(balance.Coin, balance.Total)
		}
	}
}
