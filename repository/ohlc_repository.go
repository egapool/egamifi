package repository

import "github.com/egapool/egamifi/domain"

type OhlcRepository struct {
    SqlHandler
}

func (repo *OhlcRepository) Store(ohlc domain.Ohlc) {
    _, err := repo.Execute(
        "INSERT INTO ohlcs (open,high,low,close,volume,start_time,resolution) VALUES (?,?,?,?,?,?,?)", ohlc.Open, ohlc.High, ohlc.Low,ohlc.Close,ohlc.Volume,ohlc.StartTime,ohlc.Resolution,
    )
    if err != nil {
        return
    }
}
