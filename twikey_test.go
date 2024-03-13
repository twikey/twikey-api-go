package twikey

import (
	"log"
	"os"
	"testing"
)

func newTestClient() *Client {
	c := NewClient(os.Getenv("TWIKEY_API_KEY"))
	c.BaseURL = getEnv("TWIKEY_URL", "https://api.beta.twikey.com")
	return c
}

func TestClientWithBasicFunctionalOptions(t *testing.T) {
	salt := "salty"
	baseUrl := "http://localtest.me"
	userAgent := "someUserAgent"
	apikey := "123"

	c := NewClient(apikey,
		WithBaseURL(baseUrl),
		WithSalt(salt),
		WithUserAgent(userAgent),
		WithLogger(log.Default()),
	)

	if c.APIKey != apikey {
		t.Fatalf("Expected client ApiKey to be %s but was %s", apikey, c.APIKey)
	}

	if c.BaseURL != baseUrl {
		t.Fatalf("Expected client BaseURL to be %s but was %s", baseUrl, c.BaseURL)
	}

	if c.Salt != salt {
		t.Fatalf("Expected client Salt to be %s but was %s", salt, c.Salt)
	}

	if c.UserAgent != userAgent {
		t.Fatalf("Expected client UserAgent to be %s but was %s", userAgent, c.UserAgent)
	}

	if c.Debug != log.Default() {
		t.Fatalf("Expected the default logger from log to be used")
	}
}

func TestTwikeyClient_verifyWebhook(t *testing.T) {
	c := NewClient("1234")
	err := c.VerifyWebhook("55261CBC12BF62000DE1371412EF78C874DBC46F513B078FB9FF8643B2FD4FC2", "abc=123&name=abc")
	if err != nil {
		t.Fatal(err)
	}
}
