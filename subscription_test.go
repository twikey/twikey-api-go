package twikey

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func NewMockedTestClient(server *httptest.Server) *Client {
	cl := NewClient("TEST_API_KEY")
	cl.BaseURL = server.URL

	// already set authorization token + last login
	cl.apiToken = "api-token"
	cl.lastLogin = time.Now()

	return cl
}

func AssertEquals(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper() // marks this function as a helper function -> hides caller reference.
	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

func TestHasNext(t *testing.T) {
	res := &SubscriptionListResponse{}
	res.Links.Next = "https://..."
	if !res.HasNext() {
		t.Errorf("Expected HasNext() to have returned true when a next link is available")
	}
}

func TestSubscriptionAdd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AssertEquals(t, http.MethodPost, r.Method)
		AssertEquals(t, "/creditor/subscription", r.URL.Path)
		AssertEquals(t, "my-idempotency-key", r.Header.Get("Idempotency-Key"))
		AssertEquals(t, "api-token", r.Header.Get("Authorization"))

		// check request body
		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form parameters")
			t.FailNow()
		}

		AssertEquals(t, "Monthly subscription", r.Form.Get("message"))
		AssertEquals(t, "MyRef", r.Form.Get("ref"))
		AssertEquals(t, "12.00", r.Form.Get("amount"))
		AssertEquals(t, "1m", r.Form.Get("recurrence"))
		AssertEquals(t, "2022-11-29", r.Form.Get("start"))

		json := `{
    "id": 10,
    "state": "active",
    "amount": 12.0,
    "message": "Monthly subscription",
    "ref": "MyRef",
    "plan": 0,
    "runs": 0,
    "stopAfter": 5,
    "start": "2022-11-29",
    "next": "2022-12-01",
    "recurrence": "1m",
    "mndtId": "TEST03"
}`

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	}))
	defer server.Close()
	cl := NewMockedTestClient(server)

	ctx := context.TODO()
	output, err := cl.SubscriptionAdd(ctx, &SubscriptionAddRequest{
		IdempotencyKey: "my-idempotency-key",
		MndtId:         "TESTO3",
		Message:        "Monthly subscription",
		Ref:            "MyRef",
		Amount:         12.00,
		Recurrence:     RecurrenceMonthly,
		StartDate:      "2022-11-29",
	})
	if err != nil {
		t.Errorf("Error adding subscription: %s", err)
		t.FailNow()
	}

	AssertEquals(t, 10, output.Id)
	AssertEquals(t, SubscriptionStateActive, output.State)
	AssertEquals(t, 12.0, output.Amount)
	AssertEquals(t, "Monthly subscription", output.Message)
	AssertEquals(t, "MyRef", output.Ref)
	AssertEquals(t, 5, output.StopAfter)
	AssertEquals(t, "2022-11-29", output.Start)
	AssertEquals(t, "2022-12-01", output.Next)
	AssertEquals(t, RecurrenceMonthly, output.Recurrence)
	AssertEquals(t, "TEST03", output.MndtId)
}

func TestSubscriptionUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AssertEquals(t, http.MethodPost, r.Method)
		AssertEquals(t, "/creditor/subscription/TST123/REF1", r.URL.Path)
		AssertEquals(t, "api-token", r.Header.Get("Authorization"))

		// check request body
		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form parameters")
			t.FailNow()
		}

		AssertEquals(t, "mymessage", r.Form.Get("message"))
		AssertEquals(t, "planName", r.Form.Get("plan"))
		AssertEquals(t, "10.22", r.Form.Get("amount"))

		json := `{
    "id": 10,
    "state": "active",
    "amount": 10.22,
    "message": "mymessage",
    "ref": "reference123",
    "plan": 0,
    "runs": 0,
    "stopAfter": 5,
    "start": "2022-11-29",
    "next": "2022-12-01",
    "recurrence": "1m",
    "mndtId": "TEST03"
}`

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	}))
	defer server.Close()
	cl := NewMockedTestClient(server)

	ctx := context.TODO()
	output, err := cl.SubscriptionUpdate(ctx, "TST123", "REF1", &UpdateSubscriptionRequest{
		Message: "mymessage",
		Plan:    "planName",
		Amount:  10.22,
	})
	if err != nil {
		t.Errorf("Error updating subscription: %s", err)
		t.FailNow()
	}

	AssertEquals(t, 10, output.Id)
	AssertEquals(t, SubscriptionStateActive, output.State)
	AssertEquals(t, 10.22, output.Amount)
	AssertEquals(t, "mymessage", output.Message)
	AssertEquals(t, "reference123", output.Ref)
	AssertEquals(t, 5, output.StopAfter)
	AssertEquals(t, "2022-11-29", output.Start)
	AssertEquals(t, "2022-12-01", output.Next)
	AssertEquals(t, RecurrenceMonthly, output.Recurrence)
	AssertEquals(t, "TEST03", output.MndtId)
}

func TestSubscriptionPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AssertEquals(t, http.MethodPatch, r.Method)
		AssertEquals(t, "/creditor/subscription/TST123/REF1", r.URL.Path)
		AssertEquals(t, "api-token", r.Header.Get("Authorization"))

		// check request body
		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form parameters")
			t.FailNow()
		}

		AssertEquals(t, "MNDT200", r.Form.Get("mndtId"))
		AssertEquals(t, "mymessage", r.Form.Get("message"))
		AssertEquals(t, "25.00", r.Form.Get("amount"))

		json := `{
    "id": 10,
    "state": "active",
    "amount": 25.0,
    "message": "mymessage",
    "ref": "myreference",
    "plan": 0,
    "runs": 0,
    "stopAfter": 5,
    "start": "2024-11-29",
    "next": "2024-12-01",
    "recurrence": "1m",
    "mndtId": "MNDT200"
}
`

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	}))
	defer server.Close()
	cl := NewMockedTestClient(server)

	ctx := context.TODO()
	output, err := cl.SubscriptionPatch(ctx, "TST123", "REF1", &PatchSubscriptionRequest{
		MndtId:  "MNDT200",
		Message: "mymessage",
		Amount:  25.00,
	})
	if err != nil {
		t.Errorf("Error updating subscription: %s", err)
		t.FailNow()
	}
	AssertEquals(t, 10, output.Id)
	AssertEquals(t, SubscriptionStateActive, output.State)
	AssertEquals(t, 25.0, output.Amount)
	AssertEquals(t, "mymessage", output.Message)
	AssertEquals(t, "myreference", output.Ref)
	AssertEquals(t, 5, output.StopAfter)
	AssertEquals(t, "2024-11-29", output.Start)
	AssertEquals(t, "2024-12-01", output.Next)
	AssertEquals(t, RecurrenceMonthly, output.Recurrence)
	AssertEquals(t, "MNDT200", output.MndtId)
}

func TestSubscriptionCancel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AssertEquals(t, http.MethodDelete, r.Method)
		AssertEquals(t, "/creditor/subscription/TST123/REF1", r.URL.Path)
		AssertEquals(t, "api-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	cl := NewMockedTestClient(server)

	ctx := context.TODO()
	err := cl.SubscriptionCancel(ctx, "TST123", "REF1")
	if err != nil {
		t.Errorf("Error updating subscription: %s", err)
		t.FailNow()
	}
}

func TestSubscriptionDetail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AssertEquals(t, http.MethodGet, r.Method)
		AssertEquals(t, "/creditor/subscription/TST123/REF1", r.URL.Path)
		AssertEquals(t, "api-token", r.Header.Get("Authorization"))

		json := `{
    "id": 10,
    "state": "active",
    "amount": 25.0,
    "message": "mymessage",
    "ref": "myreference",
    "plan": 0,
    "runs": 0,
    "stopAfter": 5,
    "start": "2024-11-29",
    "next": "2024-12-01",
    "recurrence": "1m",
    "mndtId": "MNDT200"
}
`

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(json))
	}))
	defer server.Close()
	cl := NewMockedTestClient(server)

	ctx := context.TODO()
	output, err := cl.SubscriptionDetail(ctx, "TST123", "REF1")
	if err != nil {
		t.Errorf("Error updating subscription: %s", err)
		t.FailNow()
	}
	AssertEquals(t, 10, output.Id)
	AssertEquals(t, SubscriptionStateActive, output.State)
	AssertEquals(t, 25.0, output.Amount)
	AssertEquals(t, "mymessage", output.Message)
	AssertEquals(t, "myreference", output.Ref)
	AssertEquals(t, 5, output.StopAfter)
	AssertEquals(t, "2024-11-29", output.Start)
	AssertEquals(t, "2024-12-01", output.Next)
	AssertEquals(t, RecurrenceMonthly, output.Recurrence)
	AssertEquals(t, "MNDT200", output.MndtId)
}

func TestSubscriptionList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AssertEquals(t, http.MethodGet, r.Method)
		AssertEquals(t, "/creditor/subscription/query", r.URL.Path)
		AssertEquals(t, "api-token", r.Header.Get("Authorization"))

		_page := r.URL.Query().Get("page")
		if _page == "" {
			// return first page
			json := `{
  "Subscriptions": [
    {
      "id": 10,
      "state": "active",
      "amount": 12.0,
      "message": "Message for customer",
      "ref": "MyRef",
      "plan": 0,
      "runs": 0,
      "stopAfter": 5,
      "start": "2022-10-29",
      "last": "2022-11-29",
      "next": "2022-12-01",
      "recurrence": "1m",
      "mndtId": "PLOPSAABO3"
    }
  ],
  "_links": {
    "next": "/creditor/subscription/query?page=1"
  }
}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(json))
		} else {
			// return empty page
			json := `{
  "Subscriptions": [],
  "_links": {}
}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(json))
		}
	}))
	defer server.Close()
	cl := NewMockedTestClient(server)
	ctx := context.TODO()

	payload := &SubscriptionListRequest{CustomerNumber: "cst1"}
	first, err := cl.SubscriptionList(ctx, payload)
	if err != nil {
		t.Errorf("Error listing subscriptions: %s", err)
		t.FailNow()
	}

	AssertEquals(t, 1, len(first.Subscriptions))
	AssertEquals(t, true, first.HasNext())

	second, err := cl.SubscriptionList(ctx, payload.NextPage())
	if err != nil {
		t.Errorf("Error listing subscriptions: %s", err)
		t.FailNow()
	}

	AssertEquals(t, 0, len(second.Subscriptions))
	AssertEquals(t, false, second.HasNext())
}

func TestSubscriptionSuspend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AssertEquals(t, http.MethodPost, r.Method)
		AssertEquals(t, "/creditor/subscription/TST123/REF1/suspend", r.URL.Path)
		AssertEquals(t, "api-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	cl := NewMockedTestClient(server)
	ctx := context.TODO()

	err := cl.SubscriptionSuspend(ctx, "TST123", "REF1")
	if err != nil {
		t.Errorf("Error updating subscription: %s", err)
		t.FailNow()
	}
}

func TestSubscriptionResume(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AssertEquals(t, http.MethodPost, r.Method)
		AssertEquals(t, "/creditor/subscription/TST123/REF1/resume", r.URL.Path)
		AssertEquals(t, "api-token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	cl := NewMockedTestClient(server)
	ctx := context.TODO()

	err := cl.SubscriptionResume(ctx, "TST123", "REF1")
	if err != nil {
		t.Errorf("Error updating subscription: %s", err)
		t.FailNow()
	}
}
