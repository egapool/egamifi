package strategy

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/private/orders"
)

// CrossOrder is taking responsibilty for operation with cross-order strategy.
type CrossOrder struct {

	// order出すかどうかの境界値
	border float64

	Short  string
	Long   string
	client *rest.Client
}

func NewCrossOrder(client *rest.Client, border float64, long, short string) *CrossOrder {
	return &CrossOrder{
		border: border,
		Long:   long,
		Short:  short,
		client: client,
	}
}

func (c *CrossOrder) ShouldEntery(diff float64) bool {
	if diff < c.border {
		return true
	}
	return false
}

func errorHandle(err error) {
	if err != nil {
		switch err.(type) {
		case *rest.APIError:
			fmt.Println(err)
		default:
			log.Fatal(err)
		}
	}
}

var (
	OrderNotfound = rest.APIError{Status: 404, Message: "Order not found"}
)

func (c *CrossOrder) UpdateTicker(diff, askPrice, size float64) {
	// 注文だすかどうかのジャッジ
	if !c.ShouldEntery(diff) {
		return
	}

	// 注文
	longOrderResponse := c.PlaceLongOrder(askPrice, size)
	if longOrderResponse.Status == "closed" {
		status2 := c.PlaceShortOrder(size)
		fmt.Println(status2)
		return
	}
	if longOrderResponse.FilledSize == 0 {
		fmt.Println("Longポジションをキャンセルします")
		r, err := c.client.CancelByID(&orders.RequestForCancelByID{OrderID: longOrderResponse.ID})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(r)
	} else if longOrderResponse.RemainingSize > 0 {
		fmt.Println("未約定部分のロングポジションをキャンセルします")
		_, err := c.client.CancelByID(&orders.RequestForCancelByID{OrderID: longOrderResponse.ID})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("部分約定の反対positionを立てます")
		r := c.PlaceShortOrder(longOrderResponse.FilledSize)
		fmt.Println(r)
		return
	}
}

func (c *CrossOrder) PlaceLongOrder(price, size float64) *orders.ResponseForOrderStatus {
	fmt.Println("Execution Long Order", c.Long, price, size)
	res, err := c.client.PlaceOrder(&orders.RequestForPlaceOrder{
		Market: c.Long,
		Side:   "buy",
		Type:   "limit",
		Size:   size,
		Price:  price,
	})
	if err != nil {
		log.Fatal(err)
	}
	wait := 0
	for {
		fmt.Println("Check Order type ID ", res.ID)
		status, err := c.client.OrderStatus(&orders.RequestForOrderStatus{
			OrderID: strconv.Itoa(res.ID),
		})
		if err != nil {
			if err.Error() == OrderNotfound.Error() {
				fmt.Println(err.Error())
				continue
			}
			log.Fatal(err)
		}
		fmt.Println(status)
		if status.Status == "closed" {
			fmt.Println("完全約定しました", c.Long)
			return status
		}
		wait++
		if wait > 10 {
			fmt.Println("約定しなかった")
			return status
		}
		time.Sleep(time.Millisecond * 100)
	}
}

func (c *CrossOrder) PlaceShortOrder(size float64) *orders.ResponseForPlaceOrder {
	fmt.Println("Execution Order", c.Short, size)
	res, err := c.client.PlaceOrder(&orders.RequestForPlaceOrder{
		Market: c.Short,
		Side:   "sell",
		Type:   "market",
		Size:   size,
	})
	if err != nil {
		log.Fatal(err)
	}
	return res

}
