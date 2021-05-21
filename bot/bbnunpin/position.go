package bbnunpin

type position struct {
	ID       int
	side     string
	size     float64
	avgPrice float64
}

func (p *position) oppositeSide() string {
	if p.side == "buy" {
		return "sell"
	} else {
		return "buy"
	}
}
