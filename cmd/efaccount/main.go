package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/egapool/egamifi/internal/account"
	"github.com/egapool/egamifi/internal/client"
)

var accountName string

func main() {
	balanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	balanceCmd.StringVar(&accountName, "account", "", "main account or your sub account name")
	balanceCmd.StringVar(&accountName, "a", "", "main account or your sub account name (shorthand)")
	if len(os.Args) < 2 {
		fmt.Println("expected 'foo' or 'bar' subcommands")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "balance":
		balanceCmd.Parse(os.Args[2:])
		client := client.NewRestClient()
		account.BalancesAll(client)
		fmt.Println("subcommand 'balance'")
		fmt.Println("  enable:", accountName)
	default:
		fmt.Println("expected 'balance' or 'bar' subcommands")
		os.Exit(1)
	}

}
