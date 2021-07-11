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
+ Go >= 1.0
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
import (
    twikey "github.com/twikey/twikey-api-go"
)

var twikeyClient = twikey.NewClient(os.Getenv("TWIKEY_API_KEY"))

var twikeyClient = &twikey.TwikeyClient{
    ApiKey:  os.Getenv("TWIKEY_API_KEY"),
    //Debug: log.Default(),
    HTTPClient: &http.Client{
        Timeout: time.Minute,
    },
}

``` 

## Documents

Invite a customer to sign a SEPA mandate using a specific behaviour template (ct) that allows you to configure
the behaviour or flow that the customer will experience. This 'ct' can be found in the template section of the settings.

```go
invite, err := twikeyClient.DocumentInvite(InviteRequest{
   ct:             yourct,
   customerNumber: "123",
   email:          "john@doe.com",
   firstname:      "John",
   lastname:       "Doe",
   l:              "en",
   address:        "Abbey road",
   city:           "Liverpool",
   zip:            "1526",
   country:        "BE",
})
if err != nil {
    t.Fatal(err)
}
fmt.println(invite.Url)
```

_After creation, the link available in invite.Url can be used to redirect the customer into the signing flow or even
send him a link through any other mechanism. Ideally you store the mandatenumber for future usage (eg. sending transactions)._


### Feed

Once signed, a webhook is sent (see below) after which you can fetch the detail through the document feed, which you can actually
think of as reading out a queue. Since it'll return you the changes since the last time you called it.

```go
err := twikeyClient.DocumentFeed(func(new Mndt) {
    fmt.println("new", new.MndtId)
}, func(update Mndt, reason AmdmntRsn) {
    fmt.println("update", update.MndtId, reason.Rsn)
}, func(mandate string, reason CxlRsn) {
    fmt.println("cancelled", mandate, reason.Rsn)
})
```

## Transactions

Send new transactions and act upon feedback from the bank.

```go
tx, err := twikeyClient.TransactionNew(TransactionRequest{
   DocumentReference: "ABC",
   Msg:               "My Transaction",
   Ref:               "My Reference",
   Amount:            10.90,
})
fmt.println("New tx", tx)
```

### Feed

```go
err := twikeyClient.TransactionFeed(func(transaction Transaction) {
    fmt.println("Transaction", transaction.Amount, transaction.BookedError, transaction.Final)
})
```

## Webhook ##

When wants to inform you about new updates about documents or payments a `webhookUrl` specified in your api settings be called.

```go
err := twikeyClient.verifyWebhook("55261CBC12BF62000DE1371412EF78C874DBC46F513B078FB9FF8643B2FD4FC2", "abc=123&name=abc")
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
