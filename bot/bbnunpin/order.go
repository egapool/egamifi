package bbnunpin

import "errors"

type Order struct {
	ID      int
	side    string
	size    float64
	purpose purpose
}

// オーダーの目的
type purpose int

const (
	InitOrder     purpose = 1
	NunpinOrder   purpose = 2
	SettleOrder   purpose = 3
	StopLossOrder purpose = 4
)

type Orders map[int]Order

func (o Orders) OneSide(side string) Orders {
	orders := Orders{}
	for _, order := range o {
		if order.side == side {
			orders[order.ID] = order
		}
	}
	return orders
}

func (o Orders) StopLossOrder() (*Order, error) {
	for _, order := range o {
		if order.purpose == StopLossOrder {
			return &order, nil
		}
	}
	return nil, errors.New("noting")
}

func (o Orders) SettleOrder() Orders {
	orders := Orders{}
	for _, order := range o {
		if order.purpose == SettleOrder {
			orders[order.ID] = order
		}
	}
	return orders
}
