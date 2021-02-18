package research

import (
	"fmt"
	"testing"
	"time"

	"github.com/egapool/egamifi/database"
	"github.com/egapool/egamifi/repository"

	"github.com/joho/godotenv"
)

func TestCompute(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}
	database.DBOpen()
	defer database.DBClose()

	repo := repository.NewOhlcRepository()
	usecase := NewDeviationRateUsecase(*repo)
	// market := "ATOM-0326"
	// fmt.Println(market)
	// usecase.SaveOhlc(market)
	q := "DEFI-0326"
	p := "DEFI-PERP"
	s, _ := time.Parse("2006-01-02 15:04:05", "2020-12-14 16:00:00")
	e, _ := time.Parse("2006-01-02 15:04:05", "2021-02-17 23:00:00")

	ret := usecase.History(q, p, s, e, "ftx")
	for _, r := range ret {
		fmt.Println(r.Time, r.DeviationRate)
	}
	// fmt.Println(ret)
}
