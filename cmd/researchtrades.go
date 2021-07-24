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
	Long:  `ex) egamifi research trades -m AXS-PERP -s "2021-07-15 23:00:00" -e "2021-07-15 23:07:00"`,
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
	filename := fmt.Sprintf("ftx-trades-%s-%s-%s.csv", tradeMarketFlag, endTime.Format("20060102150405"), startTime.Format("20060102150405"))
	filepath := dir + "/" + filename
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	jst, _ = time.LoadLocation("Asia/Tokyo")
	writer := csv.NewWriter(file)
	var lastID int = 0
	isFin := false
	var line [][]string
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
		var firstTime int64
		for _, t := range *trades {
			localTime := t.Time.In(jst)
			if firstTime == 0 {
				firstTime = localTime.Unix()
			}
			if startUnixtime >= localTime.Unix() || len(*trades) < 200 {
				isFin = true
			}
			if lastID != 0 && t.ID >= lastID {
				continue
			}
			line = append(line, []string{
				tradeMarketFlag,
				localTime.Format("2006-01-02 15:04:05.00000"),
				t.Side,
				fmt.Sprint(t.Price),
				fmt.Sprint(t.Size),
				fmt.Sprint(t.Liquidation),
				fmt.Sprint(t.ID),
			})
			lastID = t.ID
			endUnixtime = localTime.Unix()
		}
		// 1秒で200以上約定がある場合はしかたなく、次の秒にいく
		if firstTime == endUnixtime {
			endUnixtime--
		}
		writer.Flush()
		if isFin {
			break
		}
	}
	for i := len(line) - 1; i >= 0; i-- {
		writer.Write(line[i])
	}
	writer.Flush()
	fmt.Println(filepath)
}
