package template

import (
	"net/http"
	"os"
	"testing"
)

const skipIntegrationTestsMessage = `Set TEST_MINIFLUX_* environment variables to run the templates integration tests`

type integrationTestConfig struct {
	testBaseURL           string
	testAdminUsername     string
	testAdminPassword     string
	testRegularUsername   string
	testRegularPassword   string
	testFeedURL           string
	testFeedTitle         string
	testSubscriptionTitle string
	testWebsiteURL        string
}

func newIntegrationTestConfig() *integrationTestConfig {
	getDefaultEnvValues := func(key, defaultValue string) string {
		value := os.Getenv(key)
		if value == "" {
			return defaultValue
		}
		return value
	}

	return &integrationTestConfig{
		testBaseURL:           getDefaultEnvValues("TEST_MINIFLUX_BASE_URL", ""),
		testAdminUsername:     getDefaultEnvValues("TEST_MINIFLUX_ADMIN_USERNAME", ""),
		testAdminPassword:     getDefaultEnvValues("TEST_MINIFLUX_ADMIN_PASSWORD", ""),
		testRegularUsername:   getDefaultEnvValues("TEST_MINIFLUX_REGULAR_USERNAME_PREFIX", "regular_test_user"),
		testRegularPassword:   getDefaultEnvValues("TEST_MINIFLUX_REGULAR_PASSWORD", "regular_test_user_password"),
		testFeedURL:           getDefaultEnvValues("TEST_MINIFLUX_FEED_URL", "https://miniflux.app/feed.xml"),
		testFeedTitle:         getDefaultEnvValues("TEST_MINIFLUX_FEED_TITLE", "Miniflux"),
		testSubscriptionTitle: getDefaultEnvValues("TEST_MINIFLUX_SUBSCRIPTION_TITLE", "Miniflux Releases"),
		testWebsiteURL:        getDefaultEnvValues("TEST_MINIFLUX_WEBSITE_URL", "https://miniflux.app/"),
	}
}

func (c *integrationTestConfig) isConfigured() bool {
	return c.testBaseURL != "" && c.testAdminUsername != "" && c.testAdminPassword != "" && c.testFeedURL != "" && c.testFeedTitle != "" && c.testSubscriptionTitle != "" && c.testWebsiteURL != ""
}

func TestTemplateOffLine(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	resp, err := http.Get(testConfig.testBaseURL + "/offline")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("got an unexpected http code instead of 200: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}

func TestTemplateAboutUnauthenticated(t *testing.T) {
	testConfig := newIntegrationTestConfig()
	if !testConfig.isConfigured() {
		t.Skip(skipIntegrationTestsMessage)
	}

	resp, err := http.Get(testConfig.testBaseURL + "/about")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 302 {
		t.Fatalf("got an unexpected http code instead of 302: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
}
