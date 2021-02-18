/* 乖離率 */
package research

import (
	"time"

	"github.com/egapool/egamifi/domain"
	"github.com/egapool/egamifi/repository"
)

type DeviationRateUsecase struct {
	ohlcrepo repository.OhlcRepository
}

func NewDeviationRateUsecase(ohlcrepo repository.OhlcRepository) *DeviationRateUsecase {
	return &DeviationRateUsecase{
		ohlcrepo: ohlcrepo,
	}
}

func (uc *DeviationRateUsecase) History(quarter, perp string, start, end time.Time, exchanger string) domain.DeviationRates {
	var rates domain.DeviationRates
	quarter_ohlc := uc.ohlcrepo.Get(&repository.RequestForOhlcGet{
		Exchanger: exchanger,
		Market:    quarter,
		Start:     start,
		End:       end,
	})

	// TODO quarterのstartとendを使う
	perp_ohlc := uc.ohlcrepo.Get(&repository.RequestForOhlcGet{
		Exchanger: exchanger,
		Market:    perp,
		Start:     quarter_ohlc[0].StartTime,
		End:       quarter_ohlc[len(quarter_ohlc)-1].StartTime,
	})

	for i, q := range quarter_ohlc {
		rates = append(rates, domain.NewDeviationRate(q, perp_ohlc[i]))
	}
	return rates
}
