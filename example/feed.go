package main

import (
	"context"
	"fmt"
	"github.com/twikey/twikey-api-go"
	"os"
	"time"
)

func main() {
	apikey := os.Getenv("TWIKEY_API_KEY")
	apiurl := os.Getenv("TWIKEY_API_URL")
	if apikey == "" {
		fmt.Println("TWIKEY_API_KEY environment variable not set")
		os.Exit(1)
	}
	client := twikey.NewClient(apikey,
		twikey.WithBaseURL(apiurl))

	now := time.Now()
	filename_as_date := now.Format("2006-01-02T15:04:05Z")

	file, err := os.OpenFile(fmt.Sprintf("%s.csv", filename_as_date), os.O_CREATE|os.O_RDWR, 0666)

	ctx := context.Background()
	err = client.DocumentFeed(ctx,
		func(mandate *twikey.Mndt, eventTime string, eventId int64) {
			_, _ = fmt.Fprintf(file, "%s;mandate;new;%d;%s;%s\n", eventTime, eventId, mandate.MndtId, mandate.DbtrAcct)
		},
		func(originalMandateNumber string, mandate *twikey.Mndt, reason *twikey.AmdmntRsn, eventTime string, eventId int64) {
			_, _ = fmt.Fprintf(file, "%s;mandate;update;%d;%s;%s\n", eventTime, eventId, mandate.MndtId, reason.Rsn)
		},
		func(mandateNumber string, reason *twikey.CxlRsn, eventTime string, eventId int64) {
			_, _ = fmt.Fprintf(file, "%s;mandate;cancel;%d;%s;%s\n", eventTime, eventId, mandateNumber, reason.Rsn)
		},
		twikey.FeedInclude("seq"), twikey.FeedStartPosition(0))
	if err != nil {
		panic(err)
	}

	eventTime := time.Now().Format("2006-01-02T15:04:05Z")
	err = client.TransactionFeed(ctx,
		func(transaction *twikey.Transaction) {
			_, _ = fmt.Fprintf(file, "%s;transaction;update;%d;%s;%s;%s;%s\n", eventTime, transaction.Seq, transaction.BookedDate, transaction.DocumentReference, transaction.Ref, transaction.State)
		},
		twikey.FeedInclude("seq"), twikey.FeedStartPosition(0))
	if err != nil {
		panic(err)
	}

	eventTime = time.Now().Format("2006-01-02T15:04:05Z")
	err = client.PaylinkFeed(ctx,
		func(paylink *twikey.Paylink) {
			_, _ = fmt.Fprintf(file, "%s;paylink;update;%d;%d;%s;%s\n", eventTime, paylink.Seq, paylink.Id, paylink.Ref, paylink.State)
		},
		twikey.FeedInclude("seq"), twikey.FeedStartPosition(0))
	if err != nil {
		panic(err)
	}

	eventTime = time.Now().Format("2006-01-02T15:04:05Z")
	err = client.RefundFeed(ctx,
		func(refund *twikey.Refund) {
			_, _ = fmt.Fprintf(file, "%s;refund;update;%d;%s;%s;%s\n", eventTime, refund.Seq, refund.Id, refund.Ref, refund.State)
		},
		twikey.FeedInclude("seq"), twikey.FeedStartPosition(0))
	if err != nil {
		panic(err)
	}

	file.Close()
	info, err := os.Lstat(file.Name())
	if info.Size() == 0 {
		os.Remove(file.Name())
	}
}
