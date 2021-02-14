package ftx

import (
	"fmt"
	"testing"

	"github.com/egapool/egamifi/database"
	"github.com/egapool/egamifi/repository"

	"github.com/joho/godotenv"
)

func TestCompute(t *testing.T) {
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}
	database.DBOpen()
	defer database.DBClose()

	repo := repository.NewOhlcRepository()
	usecase := NewSaveOhlcUsecase(*repo)
	market := "BTC-1225"
	fmt.Println(market)
	usecase.SaveOhlc(market)
}
