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

	// // strategy := CrossOrder{
	// // 	border: 5,
	// // 	long:   "SHIT-0326",
	// // 	short:  "SHIT-PERP",
	// // 	client: client.NewSubClient("shit").Rest,
	// // }
	// client := client.NewSubClient("shit").Rest
	// _, err := client.OrderStatus(&orders.RequestForOrderStatus{
	// 	OrderID: "97878787878",
	// })
	// fmt.Println(OrderNotfound.Error())
	// if err != nil {
	// 	switch err.Error() {
	// 	case OrderNotfound.Error():
	// 		fmt.Println("404", err)
	// 	default:
	// 		fmt.Println("default", err)
	//
	// 	}
	// }
	// fmt.Println("Execution Order ")
	// strategy.UpdateTicker(10, 1933.2, 0.001)
}
