package ftx

import (
	"fmt"
	"log"
	"time"

	"github.com/egapool/egamifi/domain"
	client "github.com/egapool/egamifi/exchanger/ftx"
	"github.com/egapool/egamifi/repository"
	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/public/markets"
)

const (
	exchangerName string = "ftx"
	resolution           = 3600
)

// SaveOhlcUsecase
type SaveOhlcUsecase struct {
	ohlcrepo repository.OhlcRepository

	// infraに依存しているが今後どう影響してくるか
	client *rest.Client
}

func NewSaveOhlcUsecase(ohlcrepo repository.OhlcRepository) *SaveOhlcUsecase {
	return &SaveOhlcUsecase{
		ohlcrepo: ohlcrepo,
		client:   client.NewRestClient(),
	}
}

func (uc *SaveOhlcUsecase) SaveOhlc(market string) {
	latest := uc.ohlcrepo.Latest(market)
	var start int64
	if !latest.Empty() {
		start = latest.StartTime.Add(time.Second).Unix()
	}
	fmt.Println("Save ", market, "Start:", start)

	// 時価API取得
	// DIPに反する
	req := &markets.RequestForCandles{
		ProductCode: market,
		Resolution:  resolution,
		Limit:       5000,
		Start:       start,
	}

	candles, err := uc.client.Candles(req)
	if err != nil {
		log.Fatal(err)
	}
	var bulk []domain.Ohlc
	for _, c := range *candles {
		bulk = append(bulk, domain.Ohlc{
			Market:     market,
			Open:       c.Open,
			High:       c.High,
			Low:        c.Low,
			Close:      c.Close,
			Volume:     c.Volume,
			Resolution: resolution,
			StartTime:  c.StartTime,
			Exchanger:  exchangerName,
		})
	}
	uc.ohlcrepo.BulkStore(bulk)
}

func (uc *SaveOhlcUsecase) SaveAllOhlcs() {
	// market fetch
	futures, err := uc.client.Markets(&markets.RequestForMarkets{})
	if err != nil {
		log.Fatal(err)
	}

	// save
	for _, market := range *futures {
		if market.Underlying == "" {
			continue
		}
		uc.SaveOhlc(market.Name)
	}

}
