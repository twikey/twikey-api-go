package twikey

import (
	"testing"
)

func TestTwikeyClient_verifyWebhook(t *testing.T) {
	c := NewClient("1234")
	err := c.verifyWebhook("55261CBC12BF62000DE1371412EF78C874DBC46F513B078FB9FF8643B2FD4FC2", "abc=123&name=abc")
	if err != nil {
		t.Fatal(err)
	}
}
