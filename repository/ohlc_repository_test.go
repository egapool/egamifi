package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/egapool/egamifi/database"

	"github.com/joho/godotenv"
)

func TestCompute(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		panic(err)
	}
	database.DBOpen()
	defer database.DBClose()
	fmt.Println("Execution go test")
	repo := NewOhlcRepository()
	ohlcs := repo.Get(&RequestForOhlcGet{
		Exchanger: "ftx",
		Market:    "DEFI-PERP",
		Start:     time.Date(2021, 2, 15, 0, 0, 0, 0, time.Local),
	})
	fmt.Println(ohlcs)

	// ohlc := domain.Ohlc{
	// 	Open:      11,
	// 	High:      11,
	// 	Low:       12,
	// 	Close:     13,
	// 	Volume:    12,
	// 	StartTime: time.Now(),
	// }
	// repo.Store(ohlc)

}
