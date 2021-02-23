package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/egapool/egamifi/repository"
	"github.com/egapool/egamifi/research"
	"github.com/spf13/cobra"
)

var quarterFlag string
var perpFlag string
var startFrag string
var endFlag string
var exchangerFlag string

// researchpriceCmd represents the researchprice command
var researchdeviationCmd = &cobra.Command{
	Use:   "deviation",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if quarterFlag == "" {
			log.Fatal("quarter is required. ex. -p DEFI-0326")
		}
		if strings.Contains(quarterFlag, "PERP") {
			log.Fatal("quarter is not PERP. ex. -p DEFI-0326")
		}
		if perpFlag == "" {
			log.Fatal("perp is required. ex. -p DEFI-PERP")
		}
		if !strings.Contains(perpFlag, "PERP") {
			log.Fatal("perp must be ****-PERP. ex. -p DEFI-PERP")
		}
		if startFrag == "" {
			startFrag = time.Now().Format("2006-01-02")
		}
		if endFlag == "" {
			endFlag = time.Now().Format("2006-01-02")
		}
		if exchangerFlag == "" {
			log.Fatal("exchanger is required.")
		}
		getDeviationRate()
	},
}

func init() {
	researchCmd.AddCommand(researchdeviationCmd)
	researchdeviationCmd.Flags().StringVarP(&quarterFlag, "quarter", "q", "", "Quarter market name")
	researchdeviationCmd.Flags().StringVarP(&perpFlag, "perp", "p", "", "Perp market name")
	researchdeviationCmd.Flags().StringVarP(&startFrag, "start", "s", "", "Date string of Start. (2021-02-10)")
	researchdeviationCmd.Flags().StringVarP(&endFlag, "end", "e", "", "Date string of End. (2021-02-15)")
	researchdeviationCmd.Flags().StringVar(&exchangerFlag, "exchanger", "", "Exchagner name ftx or binance")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// researchpriceCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// researchpriceCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func getDeviationRate() {
	repo := repository.NewOhlcRepository()
	usecase := research.NewDeviationRateUsecase(*repo)
	jst, _ := time.LoadLocation("Asia/Tokyo")
	start, err := time.ParseInLocation("2006-01-02", startFrag, jst)
	if err != nil {
		log.Fatal(err)
	}
	end, err := time.ParseInLocation("2006-01-02 15:04:05", endFlag+" 23:59:59", jst)
	if err != nil {
		log.Fatal(err)
	}
	ret, err := usecase.History(quarterFlag, perpFlag, start, end, exchangerFlag)
	if err != nil {
		log.Fatal(err)
	}
	for _, r := range ret {
		fmt.Println(r.Time, r.DeviationRate)
	}
}
