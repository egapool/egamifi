package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"time"

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
	fmt.Println("start, end, avg_volume_period, against_avg_volume_rate, minimum_rate, guard_orver_bb3, minimum_candle_length, profit, pnl, fee, trade_count, win, lose, total_entry, win_rate, PF")
	avg_volume_period_list := []int{20}
	against_avg_volume_rate_list := []float64{10, 15, 20}
	minimum_rate_list := []float64{1000}
	guard_over_bb3_list := []float64{0.001, 0.003, 0.005}
	minimum_candle_length_list := []float64{1.5}

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

	execute_time := time.Now().Format("2006-01-02-150405")

	var wg sync.WaitGroup
	ch := make(chan bool, 12)
	for _, avg_volume_period := range avg_volume_period_list {
		for _, against_avg_volume_rate := range against_avg_volume_rate_list {
			for _, minimum_rate := range minimum_rate_list {
				for _, guard_over_bb3 := range guard_over_bb3_list {
					for _, minimum_candle_length := range minimum_candle_length_list {
						wg.Add(1)
						ch <- true
						inago_config := inago.NewConfig2(
							3,
							avg_volume_period,
							against_avg_volume_rate,
							minimum_rate,
							guard_over_bb3,
							minimum_candle_length,
						)
						go exec(ch, &wg, trades, inago_config, execute_time)
					}
				}
			}
		}
	}
	wg.Wait()

}
func exec(ch chan bool, wg *sync.WaitGroup, trades [][]string, inago_config inago.Config, execute_time string) {
	defer func() {
		wg.Done()
		<-ch
	}()
	logdir := fmt.Sprintf("logs/test/%s/%s/%s", testStrategy, testMarket, execute_time)
	if _, err := os.Stat(logdir); os.IsNotExist(err) {
		os.MkdirAll(logdir, 0777)
	}
	logfile := fmt.Sprintf(logdir+"/%s.log", inago_config.Serialize2())
	if err := os.Remove(logfile); err != nil {
	}

	logger := bot.NewLoggerBacktest(logfile)
	client := bot.NewTestClient()
	backtest := bot.NewBacktest(inago.NewBot(client, testMarket, inago_config, logger))
	backtest.Run(trades, true)
}
