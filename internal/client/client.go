package client

import (
	"github.com/go-numb/go-ftx/auth"
	"github.com/go-numb/go-ftx/rest"
)

const (
	API_KEY    = "QLxTwhQ-y2Iy77FxMp5zFoPub-0C2zFqFxzFGgo5"
	API_SECRET = "diUkSnRs1zHku42eju854fL-TRn-uKnML0ITUdgb"
)

func NewRestClient() *rest.Client {
	return rest.New(auth.New(API_KEY, API_SECRET))
}

func NewSubRestClient(subaccount_name string) *rest.Client {
	clientWithSubAccounts := rest.New(auth.New(
		API_KEY,
		API_SECRET,
		auth.SubAccount{
			UUID:     1,
			Nickname: subaccount_name,
		},
	))
	clientWithSubAccounts.Auth.UseSubAccountID(1)
	return clientWithSubAccounts
}
