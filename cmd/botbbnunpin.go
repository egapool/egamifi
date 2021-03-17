package cmd

import (
	"os"

	"github.com/egapool/egamifi/bot/bbnunpin"
	client "github.com/egapool/egamifi/exchanger/ftx"
	"github.com/spf13/cobra"
)

// researchpriceCmd represents the researchprice command
var botbbnunpinCmd = &cobra.Command{
	Use:   "bbn",
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
	botCmd.AddCommand(botbbnunpinCmd)
}

func runBborder() {
	client := client.NewSubRestClient(os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"), os.Getenv("FTX_SUBACCOUNT"))
	market := "BTC-PERP"
	bb := bbnunpin.NewBbNunpin(client, market)
	bb.Run()
}
