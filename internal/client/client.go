package client

import (
	"github.com/go-numb/go-ftx/auth"
	"github.com/go-numb/go-ftx/rest"
)

const (
	API_KEY    = "QLxTwhQ-y2Iy77FxMp5zFoPub-0C2zFqFxzFGgo5"
	API_SECRET = "diUkSnRs1zHku42eju854fL-TRn-uKnML0ITUdgb"
)

type Client struct {
	Rest *rest.Client
}

func NewClient() *Client {
	client := rest.New(auth.New(API_KEY, API_SECRET))
	return &Client{Rest: client}
}

func NewSubClient(subaccount_name string) *Client {
	clientWithSubAccounts := rest.New(auth.New(
		API_KEY,
		API_SECRET,
		auth.SubAccount{
			UUID:     1,
			Nickname: "shit",
		},
	))
	clientWithSubAccounts.Auth.UseSubAccountID(1)
	client := &Client{Rest: clientWithSubAccounts}
	return client
}
