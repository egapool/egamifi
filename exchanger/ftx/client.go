package ftx

import (
	"fmt"

	"github.com/go-numb/go-ftx/auth"
	"github.com/go-numb/go-ftx/rest"
)

func NewClient(api_key, api_secret, name string) *rest.Client {
	if name == "" {
		return NewRestClient(api_key, api_secret)
	} else {
		return NewSubRestClient(api_key, api_secret, name)
	}
}

func NewRestClient(api_key, api_secret string) *rest.Client {
	return rest.New(auth.New(api_key, api_secret))
}

func NewSubRestClient(api_key, api_secret, subaccount_name string) *rest.Client {
	fmt.Println(api_key, api_secret, subaccount_name)
	clientWithSubAccounts := rest.New(auth.New(
		api_key,
		api_secret,
		auth.SubAccount{
			UUID:     1,
			Nickname: subaccount_name,
		},
	))
	clientWithSubAccounts.Auth.UseSubAccountID(1)
	return clientWithSubAccounts
}
