package fundingrate

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/egapool/egamifi/internal/client"
	"github.com/go-numb/go-ftx/rest/public/futures"
	"github.com/go-numb/go-ftx/rest/public/markets"
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
			ProductCode: "DEFI-PERP",
			End:         end})
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
	limited_term_name := "0326"
	limited_term_list := map[string]markets.Market{}
	markets, err := c.Markets(&markets.RequestForMarkets{})
	for _, m := range *markets {
		if strings.Contains(m.Name, limited_term_name) {
			limited_term_list[m.Underlying] = m
		}
	}

	if err != nil {
		log.Fatal(err)
	}
	ranking := RateRanking{}
	for k, v := range ret {
		e := DailyRate{k, v}
		ranking = append(ranking, e)
	}

	sort.Sort(ranking)
	for i, entry := range ranking {
		var term_name string = ""
		var volume float64 = 0
		if market, ok := limited_term_list[strings.TrimRight(entry.market, "-PERP")]; ok {
			term_name = limited_term_name
			volume = market.VolumeUsd24H
			fmt.Printf("%d %s %.4f%%/Day %.3f%%/Month (%s) vol: %.2f\n", i, entry.market, entry.rate/float64(date)*100, entry.rate/float64(date)*100*30, term_name, volume)
		}
	}
	return ranking
}
