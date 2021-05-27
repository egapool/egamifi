/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	client "github.com/egapool/egamifi/exchanger/ftx"
	"github.com/go-numb/go-ftx/rest/public/markets"
	"github.com/spf13/cobra"
)

var marketFlag string

// researchpriceCmd represents the researchprice command
var researchmarketCmd = &cobra.Command{
	Use:   "market",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("research market called")
		checkMarket()
	},
}

func init() {
	researchCmd.AddCommand(researchmarketCmd)
	researchmarketCmd.Flags().StringVarP(&marketFlag, "market", "m", "", "Market name")
}

func checkMarket() {
	client := client.NewRestClient(os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"))
LBL1:
	for {
		markets, err := client.Markets(&markets.RequestForMarkets{})
		if err != nil {
			log.Fatal(err)
		}
		for _, m := range *markets {
			if m.Name == marketFlag {
				// notify slack
				msg := m.Name + " が新規オープンしました"
				// notification.Notify(msg, os.Getenv("SLACK_CHANNEL"), os.Getenv("SLACK_WEBHOOK"))
				fmt.Println(msg)
				break LBL1
			}
		}
		time.Sleep(time.Minute)
	}
}
