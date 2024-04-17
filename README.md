<p align="center">
  <img src="https://cdn.twikey.com/img/logo.png" height="64"/>
</p>
<h1 align="center">Twikey API client for Go</h1>

Want to allow your customers to pay in the most convenient way, then Twikey is right what you need.

Recurring or occasional payments via (Recurring) Credit Card, SEPA Direct Debit or any other payment method by bringing
your own payment service provider or by leveraging your bank contract.

Twikey offers a simple and safe multichannel solution to negotiate and collect recurring (or even occasional) payments.
Twikey has integrations with a lot of accounting and CRM packages. It is the first and only provider to operate on a
European level for Direct Debit and can work directly with all major Belgian and Dutch Banks. However you can use the
payment options of your favorite PSP to allow other customers to pay as well.

## Requirements ##

[![Go Reference](https://pkg.go.dev/badge/github.com/twikey/twikey-api-go.svg)](https://pkg.go.dev/github.com/twikey/twikey-api-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/twikey/twikey-api-go)](https://goreportcard.com/report/github.com/twikey/twikey-api-go)

To use the Twikey API client, the following things are required:

+ Get yourself a [Twikey account](https://www.twikey.com).
+ Go >= 1.6
+ Up-to-date OpenSSL (or other SSL/TLS toolkit)

## Installation ##

The easiest way to install the Twikey API client is
with go get

    $  go get -u github.com/twikey/twikey-api-go 

## How to create anything ##

The api works the same way regardless if you want to create a mandate, a transaction, an invoice or even a paylink.
the following steps should be implemented:

1. Use the Twikey API client to create or import your item.

2. Once available, our platform will send an asynchronous request to the configured webhook
   to allow the details to be retrieved. As there may be multiple items ready for you a "feed" endpoint is provided
   which acts like a queue that can be read until empty till the next time.

3. The customer returns, and should be satisfied to see that the action he took is completed.

Find our full documentation online on [api.twikey.com](https://api.twikey.com).

## Getting started ##

Initializing the Twikey API client by configuring your API key which you can find in
the [Twikey merchant interface](https://www.twikey.com).

```go
package example

import "github.com/twikey/twikey-api-go"

func main() {
   client := twikey.NewClient("YOU_API_KEY")
}

``` 

It's possible to further configure the API client if so desired using functional options. 
For example providing a custom HTTP client with a different timeout and a Custom logger
implementation.

```go
package example

import (
   "github.com/twikey/twikey-api-go"
   "log"
   "net/http"
)

func main() {
   client := twikey.NewClient("YOUR_API_KEY",
      twikey.WithHTTPClient(&http.Client{Timeout: time.Minute}),
      twikey.WithLogger(log.Default()),
   )
}

``` 

Another example, it's possible to disable logging by setting the value of the logger to `NullLogger`.
By default, the Twikey client will print logs using the default standard library logger.

```go
package example

import (
   "github.com/twikey/twikey-api-go"
   "log"
   "net/http"
)

func main() {
   client := twikey.NewClient("YOUR_API_KEY",
      twikey.WithLogger(twikey.NullLogger),
   )
}

``` 

## Documents

Invite a customer to sign a SEPA mandate using a specific behaviour template (Template) that allows you to configure
the behaviour or flow that the customer will experience. This can be found in the template section of the settings.
The extra can be used to pass in extra attributes linked to the mandate.

```go
package example

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

```

_After creation, the link available in invite.Url can be used to redirect the customer into the signing flow or even
send him a link through any other mechanism. Ideally you store the mandatenumber for future usage (eg. sending transactions)._

The DocumentSign function has a similar syntax only that it requires a method and is mostly used for interactive sessions 
where no screens are involved. See [the documentation](https://api.twikey.com) for more info.

### Feed

Once signed, a webhook is sent (see below) after which you can fetch the detail through the document feed, which you can actually
think of as reading out a queue. Since it'll return you the changes since the last time you called it.

```go
err := c.DocumentFeed(context.Background(), func(mandate *Mndt, eventTime string, eventId int64) {
    fmt.println("Document created   ", mandate.MndtId, " @ ", eventTime)
}, func(originalMandateNumber string, mandate *Mndt, reason *AmdmntRsn, eventTime string, eventId int64) {
    fmt.println("Document updated   ", originalMandateNumber, reason.Rsn, " @ ", eventTime)
}, func(mandateNumber string, reason *CxlRsn, eventTime string, eventId int64) {
    fmt.println("Document cancelled ", mandateNumber, reason.Rsn, " @ ", eventTime)
})
```

## Transactions

Send new transactions and act upon feedback from the bank.

```go
tx, err := twikeyClient.TransactionNew(context.Background(), &TransactionRequest{
   DocumentReference: "ABC",
   Msg:               "My Transaction",
   Ref:               "My Reference",
   Amount:            10.90,
})
fmt.println("New tx", tx)
```

### Feed

```go
err := twikeyClient.TransactionFeed(context.Background(), func(transaction *Transaction) {
    fmt.println("Transaction", transaction.Amount, transaction.BookedError, transaction.Final)
})
```

## Webhook ##

When wants to inform you about new updates about documents or payments a `webhookUrl` specified in your api settings be called.

```go
err := twikeyClient.VerifyWebhook("55261CBC12BF62000DE1371412EF78C874DBC46F513B078FB9FF8643B2FD4FC2", "abc=123&name=abc")
if err != nil {
    t.Fatal(err)
}
```

## API documentation ##

If you wish to learn more about our API, please visit the [Twikey Api Page](https://api.twikey.com).
API Documentation is available in English.

## Want to help us make our API client even better? ##

Want to help us make our API client even better? We
take [pull requests](https://github.com/twikey/twikey-api-python/pulls).

## Support ##

Contact: [www.twikey.com](https://www.twikey.com)
