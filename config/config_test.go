// Copyright 2019 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package config // import "miniflux.app/config"

import (
	"os"
	"testing"
)

func TestDebugModeOn(t *testing.T) {
	os.Clearenv()
	os.Setenv("DEBUG", "1")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	if !opts.HasDebugMode() {
		t.Fatalf(`Unexpected debug mode value, got "%v"`, opts.HasDebugMode())
	}
}

func TestDebugModeOff(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	if opts.HasDebugMode() {
		t.Fatalf(`Unexpected debug mode value, got "%v"`, opts.HasDebugMode())
	}
}

func TestCustomBaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	if opts.BaseURL() != "http://example.org" {
		t.Fatalf(`Unexpected base URL, got "%s"`, opts.BaseURL())
	}

	if opts.RootURL() != "http://example.org" {
		t.Fatalf(`Unexpected root URL, got "%s"`, opts.RootURL())
	}

	if opts.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, opts.BasePath())
	}
}

func TestCustomBaseURLWithTrailingSlash(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org/folder/")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	if opts.BaseURL() != "http://example.org/folder" {
		t.Fatalf(`Unexpected base URL, got "%s"`, opts.BaseURL())
	}

	if opts.RootURL() != "http://example.org" {
		t.Fatalf(`Unexpected root URL, got "%s"`, opts.RootURL())
	}

	if opts.BasePath() != "/folder" {
		t.Fatalf(`Unexpected base path, got "%s"`, opts.BasePath())
	}
}

func TestBaseURLWithoutScheme(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "example.org/folder/")

	_, err := parse()
	if err == nil {
		t.Fatalf(`Parsing must fail`)
	}
}

func TestBaseURLWithInvalidScheme(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "ftp://example.org/folder/")

	_, err := parse()
	if err == nil {
		t.Fatalf(`Parsing must fail`)
	}
}

func TestInvalidBaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example|org")

	_, err := parse()
	if err == nil {
		t.Fatalf(`Parsing must fail`)
	}
}

func TestDefaultBaseURL(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	if opts.BaseURL() != defaultBaseURL {
		t.Fatalf(`Unexpected base URL, got "%s"`, opts.BaseURL())
	}

	if opts.RootURL() != defaultBaseURL {
		t.Fatalf(`Unexpected root URL, got "%s"`, opts.RootURL())
	}

	if opts.BasePath() != "" {
		t.Fatalf(`Unexpected base path, got "%s"`, opts.BasePath())
	}
}

func TestDatabaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_URL", "foobar")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "foobar"
	result := opts.DatabaseURL()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_URL value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultDatabaseURLValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultDatabaseURL
	result := opts.DatabaseURL()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_URL value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultDatabaseMaxConnsValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultDatabaseMaxConns
	result := opts.DatabaseMaxConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MAX_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestDatabaseMaxConns(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_MAX_CONNS", "42")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := 42
	result := opts.DatabaseMaxConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MAX_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultDatabaseMinConnsValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultDatabaseMinConns
	result := opts.DatabaseMinConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MIN_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestDatabaseMinConns(t *testing.T) {
	os.Clearenv()
	os.Setenv("DATABASE_MIN_CONNS", "42")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := 42
	result := opts.DatabaseMinConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MIN_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestListenAddr(t *testing.T) {
	os.Clearenv()
	os.Setenv("LISTEN_ADDR", "foobar")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "foobar"
	result := opts.ListenAddr()

	if result != expected {
		t.Fatalf(`Unexpected LISTEN_ADDR value, got %q instead of %q`, result, expected)
	}
}

func TestListenAddrWithPortDefined(t *testing.T) {
	os.Clearenv()
	os.Setenv("PORT", "3000")
	os.Setenv("LISTEN_ADDR", "foobar")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := ":3000"
	result := opts.ListenAddr()

	if result != expected {
		t.Fatalf(`Unexpected LISTEN_ADDR value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultListenAddrValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultListenAddr
	result := opts.ListenAddr()

	if result != expected {
		t.Fatalf(`Unexpected LISTEN_ADDR value, got %q instead of %q`, result, expected)
	}
}

func TestCertFile(t *testing.T) {
	os.Clearenv()
	os.Setenv("CERT_FILE", "foobar")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "foobar"
	result := opts.CertFile()

	if result != expected {
		t.Fatalf(`Unexpected CERT_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCertFileValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultCertFile
	result := opts.CertFile()

	if result != expected {
		t.Fatalf(`Unexpected CERT_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestKeyFile(t *testing.T) {
	os.Clearenv()
	os.Setenv("KEY_FILE", "foobar")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "foobar"
	result := opts.CertKeyFile()

	if result != expected {
		t.Fatalf(`Unexpected KEY_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultKeyFileValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultKeyFile
	result := opts.CertKeyFile()

	if result != expected {
		t.Fatalf(`Unexpected KEY_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestCertDomain(t *testing.T) {
	os.Clearenv()
	os.Setenv("CERT_DOMAIN", "example.org")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "example.org"
	result := opts.CertDomain()

	if result != expected {
		t.Fatalf(`Unexpected CERT_DOMAIN value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCertDomainValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultCertDomain
	result := opts.CertDomain()

	if result != expected {
		t.Fatalf(`Unexpected CERT_DOMAIN value, got %q instead of %q`, result, expected)
	}
}

func TestCertCache(t *testing.T) {
	os.Clearenv()
	os.Setenv("CERT_CACHE", "foobar")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "foobar"
	result := opts.CertCache()

	if result != expected {
		t.Fatalf(`Unexpected CERT_CACHE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCertCacheValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultCertCache
	result := opts.CertCache()

	if result != expected {
		t.Fatalf(`Unexpected CERT_CACHE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCleanupFrequencyValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultCleanupFrequency
	result := opts.CleanupFrequency()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestCleanupFrequency(t *testing.T) {
	os.Clearenv()
	os.Setenv("CLEANUP_FREQUENCY", "42")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := 42
	result := opts.CleanupFrequency()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultWorkerPoolSizeValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultWorkerPoolSize
	result := opts.WorkerPoolSize()

	if result != expected {
		t.Fatalf(`Unexpected WORKER_POOL_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestWorkerPoolSize(t *testing.T) {
	os.Clearenv()
	os.Setenv("WORKER_POOL_SIZE", "42")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := 42
	result := opts.WorkerPoolSize()

	if result != expected {
		t.Fatalf(`Unexpected WORKER_POOL_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestDefautPollingFrequencyValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultPollingFrequency
	result := opts.PollingFrequency()

	if result != expected {
		t.Fatalf(`Unexpected POLLING_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestPollingFrequency(t *testing.T) {
	os.Clearenv()
	os.Setenv("POLLING_FREQUENCY", "42")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := 42
	result := opts.PollingFrequency()

	if result != expected {
		t.Fatalf(`Unexpected POLLING_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultBatchSizeValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultBatchSize
	result := opts.BatchSize()

	if result != expected {
		t.Fatalf(`Unexpected BATCH_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestBatchSize(t *testing.T) {
	os.Clearenv()
	os.Setenv("BATCH_SIZE", "42")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := 42
	result := opts.BatchSize()

	if result != expected {
		t.Fatalf(`Unexpected BATCH_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestOAuth2UserCreationWhenUnset(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := false
	result := opts.IsOAuth2UserCreationAllowed()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_USER_CREATION value, got %v instead of %v`, result, expected)
	}
}

func TestOAuth2UserCreationAdmin(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_USER_CREATION", "1")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := true
	result := opts.IsOAuth2UserCreationAllowed()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_USER_CREATION value, got %v instead of %v`, result, expected)
	}
}

func TestOAuth2ClientID(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_CLIENT_ID", "foobar")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "foobar"
	result := opts.OAuth2ClientID()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_ID value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ClientIDValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultOAuth2ClientID
	result := opts.OAuth2ClientID()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_ID value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2ClientSecret(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_CLIENT_SECRET", "secret")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "secret"
	result := opts.OAuth2ClientSecret()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_SECRET value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ClientSecretValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultOAuth2ClientSecret
	result := opts.OAuth2ClientSecret()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_SECRET value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2RedirectURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_REDIRECT_URL", "http://example.org")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "http://example.org"
	result := opts.OAuth2RedirectURL()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_REDIRECT_URL value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2RedirectURLValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultOAuth2RedirectURL
	result := opts.OAuth2RedirectURL()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_REDIRECT_URL value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2Provider(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_PROVIDER", "google")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "google"
	result := opts.OAuth2Provider()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_PROVIDER value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ProviderValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultOAuth2Provider
	result := opts.OAuth2Provider()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_PROVIDER value, got %q instead of %q`, result, expected)
	}
}

func TestHSTSWhenUnset(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := true
	result := opts.HasHSTS()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HSTS value, got %v instead of %v`, result, expected)
	}
}

func TestHSTS(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_HSTS", "1")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := false
	result := opts.HasHSTS()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HSTS value, got %v instead of %v`, result, expected)
	}
}

func TestDisableHTTPServiceWhenUnset(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := true
	result := opts.HasHTTPService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HTTP_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestDisableHTTPService(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_HTTP_SERVICE", "1")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := false
	result := opts.HasHTTPService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HTTP_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestDisableSchedulerServiceWhenUnset(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := true
	result := opts.HasSchedulerService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_SCHEDULER_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestDisableSchedulerService(t *testing.T) {
	os.Clearenv()
	os.Setenv("DISABLE_SCHEDULER_SERVICE", "1")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := false
	result := opts.HasSchedulerService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_SCHEDULER_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestArchiveReadDays(t *testing.T) {
	os.Clearenv()
	os.Setenv("ARCHIVE_READ_DAYS", "7")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := 7
	result := opts.ArchiveReadDays()

	if result != expected {
		t.Fatalf(`Unexpected ARCHIVE_READ_DAYS value, got %v instead of %v`, result, expected)
	}
}

func TestRunMigrationsWhenUnset(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := false
	result := opts.RunMigrations()

	if result != expected {
		t.Fatalf(`Unexpected RUN_MIGRATIONS value, got %v instead of %v`, result, expected)
	}
}

func TestRunMigrations(t *testing.T) {
	os.Clearenv()
	os.Setenv("RUN_MIGRATIONS", "yes")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := true
	result := opts.RunMigrations()

	if result != expected {
		t.Fatalf(`Unexpected RUN_MIGRATIONS value, got %v instead of %v`, result, expected)
	}
}

func TestCreateAdminWhenUnset(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := false
	result := opts.CreateAdmin()

	if result != expected {
		t.Fatalf(`Unexpected CREATE_ADMIN value, got %v instead of %v`, result, expected)
	}
}

func TestCreateAdmin(t *testing.T) {
	os.Clearenv()
	os.Setenv("CREATE_ADMIN", "true")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := true
	result := opts.CreateAdmin()

	if result != expected {
		t.Fatalf(`Unexpected CREATE_ADMIN value, got %v instead of %v`, result, expected)
	}
}

func TestPocketConsumerKeyFromEnvVariable(t *testing.T) {
	os.Clearenv()
	os.Setenv("POCKET_CONSUMER_KEY", "something")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "something"
	result := opts.PocketConsumerKey("default")

	if result != expected {
		t.Fatalf(`Unexpected POCKET_CONSUMER_KEY value, got %q instead of %q`, result, expected)
	}
}

func TestPocketConsumerKeyFromUserPrefs(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "default"
	result := opts.PocketConsumerKey("default")

	if result != expected {
		t.Fatalf(`Unexpected POCKET_CONSUMER_KEY value, got %q instead of %q`, result, expected)
	}
}

func TestProxyImages(t *testing.T) {
	os.Clearenv()
	os.Setenv("PROXY_IMAGES", "all")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := "all"
	result := opts.ProxyImages()

	if result != expected {
		t.Fatalf(`Unexpected PROXY_IMAGES value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultProxyImagesValue(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	expected := defaultProxyImages
	result := opts.ProxyImages()

	if result != expected {
		t.Fatalf(`Unexpected PROXY_IMAGES value, got %q instead of %q`, result, expected)
	}
}

func TestHTTPSOff(t *testing.T) {
	os.Clearenv()

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	if opts.HTTPS {
		t.Fatalf(`Unexpected HTTPS value, got "%v"`, opts.HTTPS)
	}
}

func TestHTTPSOn(t *testing.T) {
	os.Clearenv()
	os.Setenv("HTTPS", "on")

	opts, err := parse()
	if err != nil {
		t.Fatalf(`Parsing failure: %q`, err)
	}

	if !opts.HTTPS {
		t.Fatalf(`Unexpected HTTPS value, got "%v"`, opts.HTTPS)
	}
}
