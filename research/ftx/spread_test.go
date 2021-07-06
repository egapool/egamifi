package ftx

import (
	"testing"
)

func TestSpread(t *testing.T) {
	market := "NEO-PERP"
	usecase := NewSpreadUsecase(market)
	usecase.Run()
}
