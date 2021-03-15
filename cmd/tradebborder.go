package cmd

import (
	"os"

	client "github.com/egapool/egamifi/exchanger/ftx"
	"github.com/egapool/egamifi/trade/bborder"
	"github.com/spf13/cobra"
)

// researchpriceCmd represents the researchprice command
var tradebborderCmd = &cobra.Command{
	Use:   "bborder",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		runBborder()
	},
}

func init() {
	tradeCmd.AddCommand(tradebborderCmd)
	// researchpriceCmd.Flags().StringVarP(&priceMarketFlag, "market", "m", "", "Market name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// researchpriceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// researchpriceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runBborder() {
	client := client.NewSubRestClient(os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"), os.Getenv("FTX_SUBACCOUNT"))
	bb := bborder.NewBbOrder(client)
	market := "BTC-PERP"
	bb.Run(market)
}
