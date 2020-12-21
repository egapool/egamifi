package main

import (
	"fmt"
	"log"

	"github.com/go-numb/go-ftx/auth"
	"github.com/go-numb/go-ftx/rest"
	"github.com/go-numb/go-ftx/rest/private/funding"
)

const (
	API_KEY    = "QLxTwhQ-y2Iy77FxMp5zFoPub-0C2zFqFxzFGgo5"
	API_SECRET = "diUkSnRs1zHku42eju854fL-TRn-uKnML0ITUdgb"
)

func main() {
	client := rest.New(auth.New(API_KEY, API_SECRET))
	c := client
	// info, err := c.Information(&account.RequestForInformation{})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	var sum float64
	var end int64
	for {
		funding, err := c.Funding(&funding.Request{
			// ProductCode: "MTA-PERP",
			End: end,
			// Start:       1607029200 - (100 * 24 * 60 * 60),
		})
		if err != nil {
			log.Fatal(err)
		}
		if len(*funding) == 0 {
			break
		}
		for i, f := range *funding {
			fmt.Println(i, f.Rate, f.Payment, f.Time.Unix(), f.Time, sum)
			sum += f.Payment
			end = f.Time.Unix()
		}
		end--
	}
	fmt.Println(sum)

}
