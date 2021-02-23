/* 乖離率 */
package research

import (
	"errors"
	"fmt"
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

func (uc *DeviationRateUsecase) History(quarter, perp string, start, end time.Time, exchanger string) (domain.DeviationRates, error) {
	var rates domain.DeviationRates
	quarter_ohlc := uc.ohlcrepo.Get(&repository.RequestForOhlcGet{
		Exchanger: exchanger,
		Market:    quarter,
		Start:     start,
		End:       end,
	})
	if len(quarter_ohlc) == 0 {
		fmt.Println("No quarter ohls")
		return nil, errors.New("No result")
	}

	perp_ohlc := uc.ohlcrepo.Get(&repository.RequestForOhlcGet{
		Exchanger: exchanger,
		Market:    perp,
		Start:     quarter_ohlc[0].StartTime,
		End:       quarter_ohlc[len(quarter_ohlc)-1].StartTime,
	})

	for i, q := range quarter_ohlc {
		rates = append(rates, domain.NewDeviationRate(q, perp_ohlc[i]))
	}
	return rates, nil
}
