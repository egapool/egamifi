package ftx

import (
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
	// market := "ATOM-0326"
	// fmt.Println(market)
	// usecase.SaveOhlc(market)
	usecase.SaveAllOhlcs()
}
