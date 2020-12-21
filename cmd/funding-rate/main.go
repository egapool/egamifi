package main

import "github.com/egapool/ftx-fr/internal/fundingrate"

/*
 *
 */
func main() {
	var date int64 = 7
	fundingrate.NewLatestRanking(date)
}
