package notification

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func Notify(diff float64) {
	channel := "test"
	text := "<!channel> " + fmt.Sprintf("%g", diff) + time.Now().Format(time.UnixDate)
	jsonStr := `{"channel":"` + channel + `","text":"` + text + `"}`
	req, err := http.NewRequest(
		"POST",
		"https://hooks.slack.com/services/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",
		bytes.NewBuffer([]byte(jsonStr)),
	)
	if err != nil {
		fmt.Print(err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	defer resp.Body.Close()
}
