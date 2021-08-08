package cmd

import (
	"os"

	"github.com/egapool/egamifi/bot"
	"github.com/egapool/egamifi/bot/strategy/inago"
	"github.com/spf13/cobra"
)

var inagoMarketFlag string
var inagoLotFlag float64

// researchpriceCmd represents the researchprice command
var botInagoCmd = &cobra.Command{
	Use:   "inago",
	Short: "いなごbot",
	Long:  `いなごbot`,
	Run: func(cmd *cobra.Command, args []string) {
		runInagoBot()
	},
}

func init() {
	botCmd.AddCommand(botInagoCmd)
	botInagoCmd.Flags().StringVarP(&inagoMarketFlag, "market", "m", "AXS-PERP", "Market name")
	botInagoCmd.Flags().Float64VarP(&inagoLotFlag, "lot", "l", 3, "Market name")
}

func runInagoBot() {
	// client := client.NewSubRestClient(os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"), os.Getenv("FTX_SUBACCOUNT"))
	config := inago.NewConfig2(
		inagoLotFlag,
		20,
		5,
		4000,
		0.005,
		1,
	)
	logger := bot.NewLogger("logs.log")
	client := bot.NewClient(os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"), os.Getenv("FTX_SUBACCOUNT"))
	runner := bot.NewRunner(inago.NewBot(client, inagoMarketFlag, config, logger))
	runner.Run()
}
