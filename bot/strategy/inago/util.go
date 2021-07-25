package inago

func opSide(side string) string {
	if side == "buy" {
		return "sell"
	}
	if side == "sell" {
		return "buy"
	}
	return ""
}
