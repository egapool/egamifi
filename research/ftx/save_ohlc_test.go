package ftx

import (
	"testing"

	"github.com/egapool/egamifi/database"
	"github.com/egapool/egamifi/domain"
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
	market := "ATOM-PERP"
	// fmt.Println(market)
	usecase.SaveOhlc(market, domain.Resolution_60)
	// usecase.SaveAllOhlcs()
}
