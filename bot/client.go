package bot

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/egapool/egamifi/bot/strategy/inago"
	client "github.com/egapool/egamifi/exchanger/ftx"
	"github.com/egapool/egamifi/internal/notification"
	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/private/orders"
)

type Client struct {
	client  *rest.Client
	notifer Notifer
}

type Notifer interface {
	Notify(message string)
}

func NewClient(ftx_key, ftx_secret, ftx_subaccount string) *Client {
	var n Notifer
	if os.Getenv("SLACK_CHANNEL") == "" {
		n = notification.NewNone()
	} else {
		n = notification.NewNotifer(os.Getenv("SLACK_CHANNEL"), os.Getenv("SLACK_WEBHOOK"))
	}
	client := client.NewSubRestClient(ftx_key, ftx_secret, ftx_subaccount)
	return &Client{
		client:  client,
		notifer: n}
}

func (c *Client) MarketOrder(market string, side string, size float64, time time.Time, price float64) inago.Position {
	c.notifer.Notify("ポジション Open")
	req := &orders.RequestForPlaceOrder{
		Market: market,
		Type:   "market",
		Side:   side,
		Size:   size,
		Ioc:    true}
	o, err := c.client.PlaceOrder(req)
	if err != nil {
		c.notifer.Notify("Fail Open Order: " + err.Error())
		log.Fatal(err)
	}
	fmt.Println("Order Response from FTX: ", o)
	return inago.Position{
		// Time:  o.CreatedAt,
		Time:  time,
		Side:  o.Side,
		Size:  o.Size,
		Price: price,
	}
}

func (c *Client) Close(market string, p inago.Position, price float64) float64 {
	c.notifer.Notify("ポジション Close")
	var side string
	if p.Side == "buy" {
		side = "sell"
	} else {
		side = "buy"
	}
	req := &orders.RequestForPlaceOrder{
		Market:     market,
		Type:       "market",
		Side:       side,
		Size:       p.Size,
		ReduceOnly: true}
	fmt.Println(req)
	o, err := c.client.PlaceOrder(req)
	if err != nil {
		c.notifer.Notify("Fail Close Order: " + err.Error())
		log.Fatal(err)
	}
	fmt.Println(o)
	return price
}

type TestClient struct {
}

func NewTestClient() *TestClient {
	return &TestClient{}
}

func (t *TestClient) MarketOrder(market string, side string, size float64, time time.Time, price float64) inago.Position {
	return inago.Position{
		Time:  time,
		Side:  side,
		Size:  size,
		Price: price,
	}
}

func (t *TestClient) Close(market string, p inago.Position, price float64) float64 {
	return price
}
