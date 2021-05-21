package cmd

import (
	"os"

	"github.com/egapool/egamifi/bot/bbnunpin"
	client "github.com/egapool/egamifi/exchanger/ftx"
	"github.com/spf13/cobra"
)

var bbnunpinSizeFlag float64
var bbnunpinMarketFlag string

// researchpriceCmd represents the researchprice command
var botbbnunpinCmd = &cobra.Command{
	Use:   "bbn",
	Short: "ボリンジャーナンピンbot",
	Long:  `ボリンジャーナンピンbotです`,
	Run: func(cmd *cobra.Command, args []string) {
		runBborder()
	},
}

func init() {
	botCmd.AddCommand(botbbnunpinCmd)
	botbbnunpinCmd.Flags().Float64VarP(&bbnunpinSizeFlag, "size", "s", 0.01, "Initial size")
	botbbnunpinCmd.Flags().StringVarP(&bbnunpinMarketFlag, "market", "m", "BTC-PERP", "Market name")
}

func runBborder() {
	client := client.NewSubRestClient(os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"), os.Getenv("FTX_SUBACCOUNT"))
	bb := bbnunpin.NewBbNunpin(client, bbnunpinMarketFlag, bbnunpinSizeFlag)
	bb.Run()
}
