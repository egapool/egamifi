package fundingrate

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/go-numb/go-ftx/auth"
	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/public/futures"
	"github.com/go-numb/go-ftx/rest/public/markets"
)

// clientごと別パッケージに移す
const (
	API_KEY    = "QLxTwhQ-y2Iy77FxMp5zFoPub-0C2zFqFxzFGgo5"
	API_SECRET = "diUkSnRs1zHku42eju854fL-TRn-uKnML0ITUdgb"
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
	client := rest.New(auth.New(API_KEY, API_SECRET))
	c := client
	var t time.Time
	var pool []futures.Rate
	var end int64
	var terminate_time = time.Now().Unix() - (60 * 60 * 24 * date)
	ret := map[string]float64{}
	for {
		pool = []futures.Rate{}
		rates, err := c.Rates(&futures.RequestForRates{
			// ProductCode: "SHIT-PERP",
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
			fmt.Printf("%d %s %.4f (%s) vol: %.2f\n", i, entry.market, entry.rate/float64(date)*100, term_name, volume)
		}
	}
	return ranking
}
