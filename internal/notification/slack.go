package notification

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

type Notifer struct {
	Channel string
	Url     string
}

func NewNotifer(channel string, url string) *Notifer {
	return &Notifer{
		Channel: channel,
		Url:     url,
	}
}

func (n *Notifer) Notify(message string) {
	text := "<!channel> " + message + time.Now().Format(time.UnixDate)
	jsonStr := `{"channel":"` + n.Channel + `","text":"` + text + `"}`
	req, err := http.NewRequest(
		"POST",
		n.Url,
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
