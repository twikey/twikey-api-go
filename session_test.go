package twikey

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"
)

type TestTimeProvider struct {
	currentTime time.Time
}

func (t *TestTimeProvider) Now() time.Time {
	return t.currentTime
}

func (t *TestTimeProvider) Add(duration time.Duration) {
	t.currentTime = t.currentTime.Add(duration)
}

func TestClient_refreshTokenIfRequired(t *testing.T) {
	if os.Getenv("TWIKEY_API_KEY") == "" {
		t.Skip("No TWIKEY_API_KEY available")
	}
	ttp := TestTimeProvider{
		currentTime: time.Now(),
	}
	c := Client{
		BaseURL:   baseURLV1,
		APIKey:    os.Getenv("TWIKEY_API_KEY"),
		Salt:      "own",
		UserAgent: twikeyBaseAgent,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
		TimeProvider: &ttp,
		Debug:        log.Default(),
	}
	c.BaseURL = getEnv("TWIKEY_URL", "https://api.beta.twikey.com")

	err := c.refreshTokenIfRequired()
	if err != nil {
		t.Error(err)
	}
	firstLogin := c.lastLogin

	ttp.Add(time.Hour*23 + time.Minute*20)
	err = c.refreshTokenIfRequired()
	if err != nil {
		t.Error(err)
	}
	if firstLogin == c.lastLogin {
		t.Error("First should not equal second")
	}
}
