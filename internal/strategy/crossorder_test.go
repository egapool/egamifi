package strategy

import (
	"fmt"
	"testing"

	"github.com/egapool/egamifi/internal/client"
	"github.com/go-numb/go-ftx/rest/private/orders"
)

func TestCompute(t *testing.T) {
	fmt.Println("Execution go test")
	// strategy := CrossOrder{
	// 	border: 5,
	// 	long:   "SHIT-0326",
	// 	short:  "SHIT-PERP",
	// 	client: client.NewSubClient("shit").Rest,
	// }
	client := client.NewSubClient("shit").Rest
	_, err := client.OrderStatus(&orders.RequestForOrderStatus{
		OrderID: "97878787878",
	})
	fmt.Println(OrderNotfound.Error())
	if err != nil {
		switch err.Error() {
		case OrderNotfound.Error():
			fmt.Println("404", err)
		default:
			fmt.Println("default", err)

		}
	}
	fmt.Println("Execution Order ")
	// strategy.UpdateTicker(10, 1933.2, 0.001)
}
