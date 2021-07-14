package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	client "github.com/egapool/egamifi/exchanger/ftx"
	"github.com/go-numb/go-ftx/rest/public/markets"
	"github.com/spf13/cobra"
)

var tradeMarketFlag string
var tradeStartFlag string
var tradeEndFlag string

// researchpriceCmd represents the researchprice command
var researchtradesCmd = &cobra.Command{
	Use:   "trades",
	Short: "A brief description of your command",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("research trades called")
		if tradeStartFlag == "" {
			log.Fatal("start is required")
		}
		if tradeEndFlag == "" {
			log.Fatal("end is required")
		}
		getTrades()
	},
}

func init() {
	researchCmd.AddCommand(researchtradesCmd)
	researchtradesCmd.Flags().StringVarP(&tradeMarketFlag, "market", "m", "", "Market name")
	researchtradesCmd.Flags().StringVarP(&tradeStartFlag, "start", "s", "", "DateTime string of Start. (2021-02-10)")
	researchtradesCmd.Flags().StringVarP(&tradeEndFlag, "end", "e", "", "DateTime string of Start. (2021-02-10)")
}

func getTrades() {
	client := client.NewRestClient(os.Getenv("FTX_KEY"), os.Getenv("FTX_SECRET"))

	jst, _ := time.LoadLocation("Asia/Tokyo")
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", tradeStartFlag, jst)
	if err != nil {
		log.Fatal(err)
	}
	startUnixtime := startTime.Unix()
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", tradeEndFlag, jst)
	if err != nil {
		log.Fatal(err)
	}
	endUnixtime := endTime.Unix()

	dir := "data"
	filename := fmt.Sprintf("ftx-trades-%s-%s.csv", endTime.Format("20060102150405"), startTime.Format("20060102150405"))
	file, err := os.OpenFile(dir+"/"+filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	jst, _ = time.LoadLocation("Asia/Tokyo")
	writer := csv.NewWriter(file)
	var lastID int = 0
	isFin := false
	for {
		trades, err := client.Trades(&markets.RequestForTrades{
			ProductCode: tradeMarketFlag,
			Limit:       200,
			End:         endUnixtime + 1,
			Start:       startUnixtime,
		})
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range *trades {
			if lastID != 0 && t.ID >= lastID {
				continue
			}
			localTime := t.Time.In(jst)
			writer.Write([]string{
				tradeMarketFlag,
				localTime.Format("2006-01-02 15:04:05.00000"),
				t.Side,
				fmt.Sprint(t.Size),
				fmt.Sprint(t.Liquidation),
				fmt.Sprint(t.ID),
			})
			lastID = t.ID
			endUnixtime = localTime.Unix()
			if startUnixtime >= localTime.Unix() {
				isFin = true
			}
		}
		writer.Flush()
		if isFin {
			break
		}
	}
}
