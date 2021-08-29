package ftx

import (
	"fmt"
	"log"
	"time"

	"github.com/egapool/egamifi/internal/indicators"
	"github.com/egapool/egamifi/internal/notification"
	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/public/markets"
)

type BolingerAlert struct {
	client       *rest.Client
	market       string
	sigma        float64
	alertTime    time.Time
	notification *notification.Notifer
}

func NewBolingerAlert(client *rest.Client) *BolingerAlert {
	return &BolingerAlert{
		client:    client,
		market:    "BTC-PERP",
		sigma:     2,
		alertTime: time.Now(),
		notification: notification.NewNotifer(
			"game",
			"https://hooks.slack.com/services/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",
		),
	}
}

func (b *BolingerAlert) Run() {
	fmt.Println("ボリンジャー監視始めます")
	for {
		mf := b.fetchCandles(86400)
		last := mf[len(mf)-1:][0]
		middle, upper, lower := indicators.BollingerBands(mf, 20, 2)
		upper_price := upper[len(upper)-1:][0]
		lower_price := lower[len(lower)-1:][0]
		middle_price := middle[len(middle)-1:][0]
		volatility := (upper_price - middle_price) / 2.0

		u := upper_price - last
		l := last - lower_price
		buffer := volatility / 2.5
		if u < buffer {
			msg := fmt.Sprintf("上ボリンジャーまで `$%.0f` 切ってます", u)
			b.notification.Notify(msg)
			fmt.Println(msg)
			b.alertTime = time.Now()
			fmt.Println(b.alertTime)
		} else if l < buffer {
			msg := fmt.Sprintf("下ボリンジャーまで `$%.0f` 切ってます", l)
			b.notification.Notify(msg)
			fmt.Println(msg)
			b.alertTime = time.Now()
			fmt.Println(b.alertTime)
		}

		time.Sleep(time.Hour * 4)
	}
}

func (b *BolingerAlert) fetchCandles(resolution int) indicators.Mfloat {
	var mf indicators.Mfloat
	req := &markets.RequestForCandles{
		ProductCode: b.market,
		Resolution:  resolution,
		Limit:       121,
	}
	candles, err := b.client.Candles(req)
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range *candles {
		mf = append(mf, c.Close)
	}
	return mf
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
