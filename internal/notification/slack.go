package notification

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func Notify(message string) {
	channel := "general"
	text := "<!channel> " + message + time.Now().Format(time.UnixDate)
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
