package repository

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/egapool/egamifi/domain"
)

type OhlcRepository struct {
	*Repository
}

func NewOhlcRepository() *OhlcRepository {
	return &OhlcRepository{
		NewRepository(),
	}
}

func (repo *OhlcRepository) Latest(market string, exchanger string, resolution int) domain.Ohlc {
	var ohlc domain.Ohlc
	repo.db.Where("market = ?", market).Where("exchanger = ?", exchanger).Where("resolution = ?", resolution).Last(&ohlc)
	return ohlc
}

func (repo *OhlcRepository) Store(ohlc domain.Ohlc) {
	repo.db.Create(&domain.Ohlc{
		Market:     ohlc.Market,
		Close:      ohlc.Close,
		Open:       ohlc.Open,
		High:       ohlc.High,
		Low:        ohlc.Low,
		StartTime:  ohlc.StartTime,
		Volume:     ohlc.Volume,
		Resolution: ohlc.Resolution,
		Exchanger:  ohlc.Exchanger,
	})
}

func (repo *OhlcRepository) BulkStore(ohlcs []domain.Ohlc) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	str := "INSERT INTO ohlcs (`market`, `open`, `high`, `low`, `close`, `volume`, `resolution`, `exchanger`, `start_time`) VALUES "
	var s []string
	for i, ohlc := range ohlcs {
		var icon string
		if len(ohlcs)-1 == i {
			icon = ";"
		} else {
			icon = ","
		}
		q := fmt.Sprintf("('%s','%f','%f','%f','%f','%f','%d', '%s', '%s')%s", ohlc.Market, ohlc.Open, ohlc.High, ohlc.Low, ohlc.Close, ohlc.Volume, ohlc.Resolution, ohlc.Exchanger, ohlc.StartTime.In(jst).Format("2006-01-02 15:04:05"), icon)
		s = append(s, q)
	}

	query := strings.Join(s, "")

	repo.db.Exec(str + query)

}

type RequestForOhlcGet struct {
	Exchanger string
	Market    string
	Start     time.Time
	End       time.Time
}

func (repo *OhlcRepository) Get(r *RequestForOhlcGet) (ohlcs []domain.Ohlc) {
	if r.Exchanger == "" {
		log.Fatal("Exchanger is required.")
	}
	if r.Market == "" {
		log.Fatal("Market is required.")
	}
	query := repo.db.Where("exchanger = ?", r.Exchanger).Where("market = ?", r.Market)
	if !r.Start.IsZero() {
		query = query.Where("start_time >= ?", r.Start.Format("2006-01-02 15:04:05"))
	}
	if !r.End.IsZero() {
		query = query.Where("start_time <= ?", r.End.Format("2006-01-02 15:04:05"))
	}

	query.Find(&ohlcs)
	return
}
