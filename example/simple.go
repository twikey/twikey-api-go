package main

import (
	"context"
	"fmt"
	"github.com/twikey/twikey-api-go"
	"log"
	"os"
)

func main() {
	client := twikey.NewClient(os.Getenv("TWIKEY_API_KEY"))

	ctx := context.Background()
	invite, err := client.DocumentSign(ctx, &twikey.InviteRequest{
		Template:       "YOUR_TEMPLATE_ID",
		CustomerNumber: "123",
		Email:          "john@doe.com",
		Language:       "en",
		Lastname:       "Doe",
		Firstname:      "John",
		Address:        "Abbey Road",
		City:           "Liverpool",
		Zip:            "1562",
		Country:        "EN",
		Iban:           "GB32BARC20040198915359",
		Bic:            "GEBEBEB",
		Method:         "sms",
		Extra: map[string]string{
			"SomeKey": "VALUE",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(invite.Url)
}
