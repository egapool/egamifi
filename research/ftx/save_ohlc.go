package ftx

import (
	"log"

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
	// DIPに反する

	// 時価API取得
	candles, err := uc.client.Candles(&markets.RequestForCandles{
		ProductCode: market,
		Resolution:  resolution,
		Limit:       5000,
	})
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
