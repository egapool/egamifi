package main

import (
	"fmt"
	"log"

	"github.com/egapool/ftx-fr/internal/client"
	"github.com/go-numb/go-ftx/rest/private/funding"
)

func main() {
	c := client.NewClient().Rest

	var sum float64
	var end int64
	for {
		funding, err := c.Funding(&funding.Request{
			ProductCode: "CREAM-PERP",
			End:         end,
			// Start:       1607029200 - (100 * 24 * 60 * 60),
		})
		if err != nil {
			log.Fatal(err)
		}
		if len(*funding) == 0 {
			break
		}
		for i, f := range *funding {
			fmt.Println(i, f.Future, f.Rate, f.Payment, f.Time.Unix(), f.Time, sum)
			sum += f.Payment
			end = f.Time.Unix()
		}
		end--
	}
	fmt.Println(sum)

}
