// Copyright 2018 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

import (
	"os"
	"testing"
)

func TestGetBooleanValueWithUnsetVariable(t *testing.T) {
	os.Clearenv()
	if getBooleanValue("MY_TEST_VARIABLE") {
		t.Errorf(`Unset variables should returns false`)
	}
}

func TestGetBooleanValue(t *testing.T) {
	scenarios := map[string]bool{
		"":        false,
		"1":       true,
		"Yes":     true,
		"yes":     true,
		"True":    true,
		"true":    true,
		"on":      true,
		"false":   false,
		"off":     false,
		"invalid": false,
	}

	for input, expected := range scenarios {
		os.Clearenv()
		os.Setenv("MY_TEST_VARIABLE", input)
		result := getBooleanValue("MY_TEST_VARIABLE")
		if result != expected {
			t.Errorf(`Unexpected result for %q, got %v instead of %v`, input, result, expected)
		}
	}
}

func TestGetStringValueWithUnsetVariable(t *testing.T) {
	os.Clearenv()
	if getStringValue("MY_TEST_VARIABLE", "defaultValue") != "defaultValue" {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestGetStringValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("MY_TEST_VARIABLE", "test")
	if getStringValue("MY_TEST_VARIABLE", "defaultValue") != "test" {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestGetIntValueWithUnsetVariable(t *testing.T) {
	os.Clearenv()
	if getIntValue("MY_TEST_VARIABLE", 42) != 42 {
		t.Errorf(`Unset variables should returns the default value`)
	}
}

func TestGetIntValueWithInvalidInput(t *testing.T) {
	os.Clearenv()
	os.Setenv("MY_TEST_VARIABLE", "invalid integer")
	if getIntValue("MY_TEST_VARIABLE", 42) != 42 {
		t.Errorf(`Invalid integer should returns the default value`)
	}
}

func TestGetIntValue(t *testing.T) {
	os.Clearenv()
	os.Setenv("MY_TEST_VARIABLE", "2018")
	if getIntValue("MY_TEST_VARIABLE", 42) != 2018 {
		t.Errorf(`Defined variables should returns the specified value`)
	}
}

func TestDebugModeOn(t *testing.T) {
	os.Clearenv()
	os.Setenv("DEBUG", "1")
	cfg := NewConfig()

	if !cfg.HasDebugMode() {
		t.Fatalf(`Unexpected debug mode value, got "%v"`, cfg.HasDebugMode())
	}
}

func TestDebugModeOff(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()

	if cfg.HasDebugMode() {
		t.Fatalf(`Unexpected debug mode value, got "%v"`, cfg.HasDebugMode())
	}
}

func TestCustomBaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://example.org" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://example.org" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestCustomBaseURLWithTrailingSlash(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org/folder/")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://example.org/folder" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://example.org" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "/folder" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestBaseURLWithoutScheme(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "example.org/folder/")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://localhost" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://localhost" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestBaseURLWithInvalidScheme(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "ftp://example.org/folder/")
	cfg := NewConfig()

	if cfg.BaseURL() != "http://localhost" {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != "http://localhost" {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestInvalidBaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example|org")
	cfg := NewConfig()

	if cfg.BaseURL() != defaultBaseURL {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != defaultBaseURL {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestDefaultBaseURL(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()

	if cfg.BaseURL() != defaultBaseURL {
		t.Fatalf(`Unexpected base URL, got "%s"`, cfg.BaseURL())
	}

	if cfg.RootURL() != defaultBaseURL {
		t.Fatalf(`Unexpected root URL, got "%s"`, cfg.RootURL())
	}

	if cfg.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, cfg.BasePath())
	}
}

func TestDatabaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "foobar")

	cfg := NewConfig()
	expected := "foobar"
	result := cfg.DatabaseURL()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_URL value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultDatabaseURLValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.DatabaseURL()
	expected := defaultDatabaseURL

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_URL value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultDatabaseMaxConnsValue(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := defaultDatabaseMaxConns
	result := cfg.DatabaseMaxConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MAX_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestDeatabaseMaxConns(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_MAX_CONNS", "42")

	cfg := NewConfig()
	expected := 42
	result := cfg.DatabaseMaxConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MAX_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultDatabaseMinConnsValue(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := defaultDatabaseMinConns
	result := cfg.DatabaseMinConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MIN_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestDatabaseMinConns(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_MIN_CONNS", "42")

	cfg := NewConfig()
	expected := 42
	result := cfg.DatabaseMinConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MIN_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestListenAddr(t *testing.T) {
	os.Clearenv()
	os.Setenv("LISTEN_ADDR", "foobar")

	cfg := NewConfig()
	expected := "foobar"
	result := cfg.ListenAddr()

	if result != expected {
		t.Fatalf(`Unexpected LISTEN_ADDR value, got %q instead of %q`, result, expected)
	}
}

func TestListenAddrWithPortDefined(t *testing.T) {
	os.Clearenv()
	os.Setenv("PORT", "3000")
	os.Setenv("LISTEN_ADDR", "foobar")

	cfg := NewConfig()
	expected := ":3000"
	result := cfg.ListenAddr()

	if result != expected {
		t.Fatalf(`Unexpected LISTEN_ADDR value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultListenAddrValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.ListenAddr()
	expected := defaultListenAddr

	if result != expected {
		t.Fatalf(`Unexpected LISTEN_ADDR value, got %q instead of %q`, result, expected)
	}
}

func TestCertFile(t *testing.T) {
	os.Clearenv()
	os.Setenv("CERT_FILE", "foobar")

	cfg := NewConfig()
	expected := "foobar"
	result := cfg.CertFile()

	if result != expected {
		t.Fatalf(`Unexpected CERT_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCertFileValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.CertFile()
	expected := defaultCertFile

	if result != expected {
		t.Fatalf(`Unexpected CERT_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestKeyFile(t *testing.T) {
	os.Clearenv()
	os.Setenv("KEY_FILE", "foobar")

	cfg := NewConfig()
	expected := "foobar"
	result := cfg.KeyFile()

	if result != expected {
		t.Fatalf(`Unexpected KEY_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultKeyFileValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.KeyFile()
	expected := defaultKeyFile

	if result != expected {
		t.Fatalf(`Unexpected KEY_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestCertDomain(t *testing.T) {
	os.Clearenv()
	os.Setenv("CERT_DOMAIN", "example.org")

	cfg := NewConfig()
	expected := "example.org"
	result := cfg.CertDomain()

	if result != expected {
		t.Fatalf(`Unexpected CERT_DOMAIN value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCertDomainValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.CertDomain()
	expected := defaultCertDomain

	if result != expected {
		t.Fatalf(`Unexpected CERT_DOMAIN value, got %q instead of %q`, result, expected)
	}
}

func TestCertCache(t *testing.T) {
	os.Clearenv()
	os.Setenv("CERT_CACHE", "foobar")

	cfg := NewConfig()
	expected := "foobar"
	result := cfg.CertCache()

	if result != expected {
		t.Fatalf(`Unexpected CERT_CACHE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCertCacheValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.CertCache()
	expected := defaultCertCache

	if result != expected {
		t.Fatalf(`Unexpected CERT_CACHE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCleanupFrequencyValue(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := defaultCleanupFrequency
	result := cfg.CleanupFrequency()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestCleanupFrequency(t *testing.T) {
	os.Clearenv()
	os.Setenv("CLEANUP_FREQUENCY", "42")

	cfg := NewConfig()
	expected := 42
	result := cfg.CleanupFrequency()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultCacheFrequencyValue(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := defaultCacheFrequency
	result := cfg.CacheFrequency()

	if result != expected {
		t.Fatalf(`Unexpected CACHE_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestCacheFrequency(t *testing.T) {
	os.Clearenv()
	os.Setenv("CACHE_FREQUENCY", "42")

	cfg := NewConfig()
	expected := 42
	result := cfg.CacheFrequency()

	if result != expected {
		t.Fatalf(`Unexpected CACHE_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultWorkerPoolSizeValue(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := defaultWorkerPoolSize
	result := cfg.WorkerPoolSize()

	if result != expected {
		t.Fatalf(`Unexpected WORKER_POOL_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestWorkerPoolSize(t *testing.T) {
	os.Clearenv()
	os.Setenv("WORKER_POOL_SIZE", "42")

	cfg := NewConfig()
	expected := 42
	result := cfg.WorkerPoolSize()

	if result != expected {
		t.Fatalf(`Unexpected WORKER_POOL_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestDefautPollingFrequencyValue(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := defaultPollingFrequency
	result := cfg.PollingFrequency()

	if result != expected {
		t.Fatalf(`Unexpected POLLING_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestPollingFrequency(t *testing.T) {
	os.Clearenv()
	os.Setenv("POLLING_FREQUENCY", "42")

	cfg := NewConfig()
	expected := 42
	result := cfg.PollingFrequency()

	if result != expected {
		t.Fatalf(`Unexpected POLLING_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultBatchSizeValue(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := defaultBatchSize
	result := cfg.BatchSize()

	if result != expected {
		t.Fatalf(`Unexpected BATCH_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestBatchSize(t *testing.T) {
	os.Clearenv()
	os.Setenv("BATCH_SIZE", "42")

	cfg := NewConfig()
	expected := 42
	result := cfg.BatchSize()

	if result != expected {
		t.Fatalf(`Unexpected BATCH_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestOAuth2UserCreationWhenUnset(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := false
	result := cfg.IsOAuth2UserCreationAllowed()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_USER_CREATION value, got %v instead of %v`, result, expected)
	}
}

func TestOAuth2UserCreationAdmin(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_USER_CREATION", "1")

	cfg := NewConfig()
	expected := true
	result := cfg.IsOAuth2UserCreationAllowed()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_USER_CREATION value, got %v instead of %v`, result, expected)
	}
}

func TestOAuth2ClientID(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_CLIENT_ID", "foobar")

	cfg := NewConfig()
	expected := "foobar"
	result := cfg.OAuth2ClientID()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_ID value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ClientIDValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.OAuth2ClientID()
	expected := defaultOAuth2ClientID

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_ID value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2ClientSecret(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_CLIENT_SECRET", "secret")

	cfg := NewConfig()
	expected := "secret"
	result := cfg.OAuth2ClientSecret()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_SECRET value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ClientSecretValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.OAuth2ClientSecret()
	expected := defaultOAuth2ClientSecret

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_SECRET value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2RedirectURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_REDIRECT_URL", "http://example.org")

	cfg := NewConfig()
	expected := "http://example.org"
	result := cfg.OAuth2RedirectURL()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_REDIRECT_URL value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2RedirectURLValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.OAuth2RedirectURL()
	expected := defaultOAuth2RedirectURL

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_REDIRECT_URL value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2Provider(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_PROVIDER", "google")

	cfg := NewConfig()
	expected := "google"
	result := cfg.OAuth2Provider()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_PROVIDER value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ProviderValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.OAuth2Provider()
	expected := defaultOAuth2Provider

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_PROVIDER value, got %q instead of %q`, result, expected)
	}
}

func TestHSTSWhenUnset(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := true
	result := cfg.HasHSTS()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HSTS value, got %v instead of %v`, result, expected)
	}
}

func TestHSTS(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_HSTS", "1")

	cfg := NewConfig()
	expected := false
	result := cfg.HasHSTS()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HSTS value, got %v instead of %v`, result, expected)
	}
}

func TestDisableHTTPServiceWhenUnset(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := true
	result := cfg.HasHTTPService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HTTP_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestDisableHTTPService(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_HTTP_SERVICE", "1")

	cfg := NewConfig()
	expected := false
	result := cfg.HasHTTPService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HTTP_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestEnableCacheServiceWhenUnset(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := true
	result := cfg.HasCacheService()

	if result != expected {
		t.Fatalf(`Unexpected HasCacheService() value, got %v instead of %v`, result, expected)
	}
}

func TestDisableCacheService(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_CACHE_SERVICE", "1")

	cfg := NewConfig()
	expected := false
	result := cfg.HasCacheService()

	if result != expected {
		t.Fatalf(`Unexpected HasCacheService() value, got %v instead of %v`, result, expected)
	}
}

func TestDisableCacheServiceWhenHTTPServiceDisabled(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_HTTP_SERVICE", "1")

	cfg := NewConfig()
	expected := false
	result := cfg.HasCacheService()

	if result != expected {
		t.Fatalf(`Unexpected HasCacheService() value, got %v instead of %v`, result, expected)
	}
}

func TestDisableSchedulerServiceWhenUnset(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := true
	result := cfg.HasSchedulerService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_SCHEDULER_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestDisableSchedulerService(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_SCHEDULER_SERVICE", "1")

	cfg := NewConfig()
	expected := false
	result := cfg.HasSchedulerService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_SCHEDULER_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestArchiveReadDays(t *testing.T) {
	os.Clearenv()
	os.Setenv("ARCHIVE_READ_DAYS", "7")

	cfg := NewConfig()
	expected := 7
	result := cfg.ArchiveReadDays()

	if result != expected {
		t.Fatalf(`Unexpected ARCHIVE_READ_DAYS value, got %v instead of %v`, result, expected)
	}
}

func TestRunMigrationsWhenUnset(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := false
	result := cfg.RunMigrations()

	if result != expected {
		t.Fatalf(`Unexpected RUN_MIGRATIONS value, got %v instead of %v`, result, expected)
	}
}

func TestRunMigrations(t *testing.T) {
	os.Clearenv()
	os.Setenv("RUN_MIGRATIONS", "yes")

	cfg := NewConfig()
	expected := true
	result := cfg.RunMigrations()

	if result != expected {
		t.Fatalf(`Unexpected RUN_MIGRATIONS value, got %v instead of %v`, result, expected)
	}
}

func TestCreateAdminWhenUnset(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := false
	result := cfg.CreateAdmin()

	if result != expected {
		t.Fatalf(`Unexpected CREATE_ADMIN value, got %v instead of %v`, result, expected)
	}
}

func TestCreateAdmin(t *testing.T) {
	os.Clearenv()
	os.Setenv("CREATE_ADMIN", "true")

	cfg := NewConfig()
	expected := true
	result := cfg.CreateAdmin()

	if result != expected {
		t.Fatalf(`Unexpected CREATE_ADMIN value, got %v instead of %v`, result, expected)
	}
}

func TestPocketConsumerKeyFromEnvVariable(t *testing.T) {
	os.Clearenv()
	os.Setenv("POCKET_CONSUMER_KEY", "something")

	cfg := NewConfig()
	expected := "something"
	result := cfg.PocketConsumerKey("default")

	if result != expected {
		t.Fatalf(`Unexpected POCKET_CONSUMER_KEY value, got %q instead of %q`, result, expected)
	}
}

func TestPocketConsumerKeyFromUserPrefs(t *testing.T) {
	os.Clearenv()

	cfg := NewConfig()
	expected := "default"
	result := cfg.PocketConsumerKey("default")

	if result != expected {
		t.Fatalf(`Unexpected POCKET_CONSUMER_KEY value, got %q instead of %q`, result, expected)
	}
}

func TestProxyImages(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	cfg := NewConfig()
	expected := "all"
	result := cfg.ProxyImages()

	if result != expected {
		t.Fatalf(`Unexpected PROXY_IMAGES value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultProxyImagesValue(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()
	result := cfg.ProxyImages()
	expected := defaultProxyImages

	if result != expected {
		t.Fatalf(`Unexpected PROXY_IMAGES value, got %q instead of %q`, result, expected)
	}
}

func TestHTTPSOff(t *testing.T) {
	os.Clearenv()
	cfg := NewConfig()

	if cfg.IsHTTPS {
		t.Fatalf(`Unexpected HTTPS value, got "%v"`, cfg.IsHTTPS)
	}
}

func TestHTTPSOn(t *testing.T) {
	os.Clearenv()
	os.Setenv("HTTPS", "on")
	cfg := NewConfig()

	if !cfg.IsHTTPS {
		t.Fatalf(`Unexpected HTTPS value, got "%v"`, cfg.IsHTTPS)
	}
}
