package bbnunpin

type position struct {
	side      string
	size      float64
	avgPrice  float64
	initPrice float64
	settleCnt int
}

func (p *position) oppositeSide() string {
	if p.side == "buy" {
		return "sell"
	} else {
		return "buy"
	}
}

func (p *position) stopLossPrice(priceRange float64) float64 {
	if p.side == "buy" {
		return p.avgPrice - priceRange
	} else {
		return p.avgPrice + priceRange
	}
}
