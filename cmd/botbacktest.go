package cmd

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/egapool/egamifi/bot"
	"github.com/egapool/egamifi/bot/strategy/inago"
	"github.com/spf13/cobra"
)

var testStrategy string
var testDataSource string
var testMarket string

var botbacktestCmd = &cobra.Command{
	Use:   "test",
	Short: "Back test",
	Long: `Back test
    ex) egamifi bot test -strategy inago -d ./data/ftx-trades-AXS-PERP-20210718220000-20210717220000.csv`,
	Run: func(cmd *cobra.Command, args []string) {
		runBacktest()
	},
}

func init() {
	botCmd.AddCommand(botbacktestCmd)
	botbacktestCmd.Flags().StringVarP(&testMarket, "market", "m", "AXS-PERP", "Market Sinbol")
	botbacktestCmd.Flags().StringVarP(&testStrategy, "strategy", "s", "inago", "Strategy")
	botbacktestCmd.Flags().StringVarP(&testDataSource, "data", "d", "", "Data source path")
}

func runBacktest() {
	switch testStrategy {
	case "inago":
		runInago()
	}
}

func runInago() {
	fmt.Println("start, end, triger_volume, scope, settle_term, settle_range, price_ratio, reverse, profit, pnl, fee, trade_count, win, lose, total_entry, win_rate, PF")
	volume_triger_list := []float64{1400, 1450, 1500, 1550, 1600}
	scope_list := []int64{20, 40, 60}
	settle_term_list := []int64{30, 50}
	reverse_list := []bool{false}
	settle_range_list := []float64{0.02, 0.025, 0.03, 0.035}
	price_ratio_list := []float64{15, 25, 30, 40, 50}

	// goroutineで使うためにメモリに読み込み
	// AXSの場合1日で〜20MB
	file, err := os.Open(testDataSource)
	if err != nil {
		panic(err)
	}
	reader := csv.NewReader(file)
	var line []string
	var trades [][]string
	for {
		line, err = reader.Read()
		if err != nil {
			break
		}
		trades = append(trades, line)
	}
	file.Close()

	ch := make(chan bool, 12)
	for _, volumeTriger := range volume_triger_list {
		for _, scope := range scope_list {
			for _, settleTerm := range settle_term_list {
				for _, reverse := range reverse_list {
					for _, settle_range := range settle_range_list {
						for _, price_ratio := range price_ratio_list {
							ch <- true
							inago_config := inago.NewConfig(
								scope,
								volumeTriger,
								settleTerm,
								settle_range,
								price_ratio,
								reverse)
							go exec(ch, trades, inago_config)
						}
					}
				}
			}
		}
	}

}
func exec(ch chan bool, trades [][]string, inago_config inago.Config) {
	defer func() {
		<-ch
	}()
	backtest := bot.NewBacktest(inago.NewBot(testMarket, inago_config))
	backtest.Run(trades, true)
}
