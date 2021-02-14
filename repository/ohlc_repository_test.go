package repository

import (
	"fmt"
	"testing"
	"time"

	"github.com/egapool/egamifi/database"
	"github.com/egapool/egamifi/domain"

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
	ohlc := domain.Ohlc{
		Open:      11,
		High:      11,
		Low:       12,
		Close:     13,
		Volume:    12,
		StartTime: time.Now(),
	}
	repo.Store(ohlc)

}
