package bbnunpin

type Order struct {
	ID   int
	side string
	size float64
}

type Orders map[int]Order

func (o Orders) OneSide(side string) Orders {
	var orders Orders
	for _, order := range o {
		if order.side == side {
			orders[order.ID] = order
		}
	}
	return orders
}
