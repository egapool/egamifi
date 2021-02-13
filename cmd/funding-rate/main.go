package main

import "github.com/egapool/egamifi/internal/fundingrate"

/*
 *
 */
func main() {
	var date int64 = 13
	fundingrate.NewLatestRanking(date)
}
