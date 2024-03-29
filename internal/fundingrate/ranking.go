package fundingrate

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/egapool/egamifi/internal/client"
	"github.com/go-numb/go-ftx/rest/public/futures"
)

type RateRanking []DailyRate

type DailyRate struct {
	market string
	rate   float64
}

func (l RateRanking) Len() int {
	return len(l)
}

func (l RateRanking) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l RateRanking) Less(i, j int) bool {
	if l[i].rate == l[j].rate {
		return (l[i].market > l[j].market)
	} else {
		return (l[i].rate > l[j].rate)
	}
}

// TODO volume合算していない
func NewLatestRanking(date int64) RateRanking {
	c := client.NewSubRestClient("shit")
	var t time.Time
	var pool []futures.Rate
	var end int64
	var terminate_time = time.Now().Unix() - (60 * 60 * 24 * date)
	ret := map[string]float64{}
	for {
		pool = []futures.Rate{}
		rates, err := c.Rates(&futures.RequestForRates{
			// ProductCode: "DEFI-PERP",
			End: end})
		if err != nil {
			log.Fatal(err)
		}

		for i, r := range *rates {
			if t.Unix() != r.Time.Unix() {
				for _, p := range pool {
					ret[p.Future] = ret[p.Future] + p.Rate
					fmt.Printf("%03d, %s, %f, %s, %f\n", i, p.Future, p.Rate, p.Time, ret[p.Future])
					end = p.Time.Unix() - 1
				}
				pool = []futures.Rate{}
				if t.Unix() > 0 && end < terminate_time {
					break
				}
			}
			pool = append(pool, r)
			t = r.Time
		}
		if end < terminate_time {
			break
		}
	}
	ranking := RateRanking{}
	for k, v := range ret {
		e := DailyRate{k, v}
		ranking = append(ranking, e)
	}

	sort.Sort(ranking)
	for i, entry := range ranking {
		fmt.Printf("%d %s %.4f%%/Day %.3f%%/Month  vol: \n", i, entry.market, entry.rate/float64(date)*100, entry.rate/float64(date)*100*30)
	}
	return ranking
}
