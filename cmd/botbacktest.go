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
	fmt.Println("start, end, triger_volume, scope, settle_term, reverse, profit, pnl, fee, long_count, short_count, win, lose, total_entry, rate")
	volume_triger_list := []float64{50000, 100000, 150000, 200000, 300000, 400000, 500000}
	scope_list := []int64{10, 20, 30, 40, 50, 60, 100}
	settle_term_list := []int64{10, 20, 30, 40, 50, 60, 100}
	reverse_list := []bool{true, false}

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

	ch := make(chan bool, 5)
	for _, volumeTriger := range volume_triger_list {
		for _, scope := range scope_list {
			for _, settleTerm := range settle_term_list {
				for _, reverse := range reverse_list {
					ch <- true
					inago_config := inago.NewConfig(scope, volumeTriger, settleTerm, reverse)
					go exec(ch, trades, inago_config)
				}
			}
		}
	}

}
func exec(ch chan bool, trades [][]string, inago_config inago.Config) {
	defer func() {
		<-ch
	}()
	backtest := bot.NewBacktest(inago.NewBot("AXS-PERP", inago_config))
	backtest.Run(trades, true)
}
