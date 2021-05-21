package cmd

import (
	"os"

	client "github.com/egapool/egamifi/exchanger/ftx"
	"github.com/egapool/egamifi/research/ftx"
	"github.com/spf13/cobra"
)

// researchpriceCmd represents the researchprice command
var researchBbaCmd = &cobra.Command{
	Use:   "bba",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func init() {
	researchCmd.AddCommand(researchBbaCmd)
}

func run() {
	client := client.NewRestClient(os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"))
	bba := ftx.NewBolingerAlert(client)
	bba.Run()
}
