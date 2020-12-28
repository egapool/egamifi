package main

import "github.com/egapool/ftx-fr/internal/fundingrate"

/*
 *
 */
func main() {
	var date int64 = 13
	fundingrate.NewLatestRanking(date)
}
