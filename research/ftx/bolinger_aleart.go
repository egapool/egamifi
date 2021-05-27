package ftx

import (
	"fmt"
	"log"
	"time"

	"github.com/egapool/egamifi/internal/indicators"
	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/public/markets"
)

type BolingerAlert struct {
	client    *rest.Client
	market    string
	sigma     float64
	alertTime time.Time
}

func NewBolingerAlert(client *rest.Client) *BolingerAlert {
	return &BolingerAlert{
		client:    client,
		market:    "BTC-PERP",
		sigma:     2,
		alertTime: time.Now(),
	}
}

func (b *BolingerAlert) Run() {
	fmt.Println("ボリンジャー監視始めます")
	for {
		mf := b.fetch1800Candles()
		last := mf[len(mf)-1:][0]
		_, upper, lower := indicators.BollingerBands(mf, 20, 2)
		upper_price := upper[len(upper)-1:][0]
		lower_price := lower[len(lower)-1:][0]

		u := upper_price - last
		l := last - lower_price
		border := 500.0
		if u < border {
			msg := fmt.Sprintf("上ボリンジャーまで `$%f` 切ってます", u)
			// notification.Notify(msg, "game", "https://hooks.slack.com/services/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")
			fmt.Println(msg)
			b.alertTime = time.Now()
			fmt.Println(b.alertTime)
		} else if l < border {
			msg := fmt.Sprintf("下ボリンジャーまで `$%f` 切ってます", l)
			// notification.Notify(msg, "game", "https://hooks.slack.com/services/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")
			fmt.Println(msg)
			b.alertTime = time.Now()
			fmt.Println(b.alertTime)
		}

		time.Sleep(time.Minute * 5)
	}
}

func (b *BolingerAlert) fetch1800Candles() indicators.Mfloat {
	var mf indicators.Mfloat

	req := &markets.RequestForCandles{
		ProductCode: b.market,
		Resolution:  900,
		Limit:       121,
	}
	candles, err := b.client.Candles(req)
	if err != nil {
		log.Fatal(err)
	}
	for i, c := range *candles {
		if i == 0 || i%2 == 0 {
			mf = append(mf, c.Close)
		}
	}
	return mf

}
