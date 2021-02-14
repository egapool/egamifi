package repository

import "github.com/egapool/egamifi/domain"

type OhlcRepository struct {
	*Repository
}

func NewOhlcRepository() *OhlcRepository {
	return &OhlcRepository{
		NewRepository(),
	}
}

func (repo *OhlcRepository) Store(ohlc domain.Ohlc) {
	repo.db.Create(&domain.Ohlc{
		Close:      ohlc.Close,
		Open:       ohlc.Open,
		High:       ohlc.High,
		Low:        ohlc.Low,
		StartTime:  ohlc.StartTime,
		Volume:     ohlc.Volume,
		Resolution: ohlc.Resolution,
	})
}
