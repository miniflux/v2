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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	if !opts.HasDebugMode() {
		t.Fatalf(`Unexpected debug mode value, got "%v"`, opts.HasDebugMode())
	}
}

func TestDebugModeOff(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	if opts.HasDebugMode() {
		t.Fatalf(`Unexpected debug mode value, got "%v"`, opts.HasDebugMode())
	}
}

func TestCustomBaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example.org")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	_, err := NewParser().ParseEnvironmentVariables()
	if err == nil {
		t.Fatalf(`Parsing must fail`)
	}
}

func TestBaseURLWithInvalidScheme(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "ftp://example.org/folder/")

	_, err := NewParser().ParseEnvironmentVariables()
	if err == nil {
		t.Fatalf(`Parsing must fail`)
	}
}

func TestInvalidBaseURL(t *testing.T) {
	os.Clearenv()
	os.Setenv("BASE_URL", "http://example|org")

	_, err := NewParser().ParseEnvironmentVariables()
	if err == nil {
		t.Fatalf(`Parsing must fail`)
	}
}

func TestDefaultBaseURL(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "foobar"
	result := opts.DatabaseURL()

	if result != expected {
		t.Errorf(`Unexpected DATABASE_URL value, got %q instead of %q`, result, expected)
	}

	if opts.IsDefaultDatabaseURL() {
		t.Errorf(`This is not the default database URL and it should returns false`)
	}
}

func TestDefaultDatabaseURLValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultDatabaseURL
	result := opts.DatabaseURL()

	if result != expected {
		t.Errorf(`Unexpected DATABASE_URL value, got %q instead of %q`, result, expected)
	}

	if !opts.IsDefaultDatabaseURL() {
		t.Errorf(`This is the default database URL and it should returns true`)
	}
}

func TestDefaultDatabaseMaxConnsValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 42
	result := opts.DatabaseMaxConns()

	if result != expected {
		t.Fatalf(`Unexpected DATABASE_MAX_CONNS value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultDatabaseMinConnsValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := ":3000"
	result := opts.ListenAddr()

	if result != expected {
		t.Fatalf(`Unexpected LISTEN_ADDR value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultListenAddrValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "foobar"
	result := opts.CertFile()

	if result != expected {
		t.Fatalf(`Unexpected CERT_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCertFileValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "foobar"
	result := opts.CertKeyFile()

	if result != expected {
		t.Fatalf(`Unexpected KEY_FILE value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultKeyFileValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "example.org"
	result := opts.CertDomain()

	if result != expected {
		t.Fatalf(`Unexpected CERT_DOMAIN value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCertDomainValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultCertDomain
	result := opts.CertDomain()

	if result != expected {
		t.Fatalf(`Unexpected CERT_DOMAIN value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultCleanupFrequencyHoursValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultCleanupFrequencyHours
	result := opts.CleanupFrequencyHours()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_FREQUENCY_HOURS value, got %v instead of %v`, result, expected)
	}
}

func TestCleanupFrequencyHours(t *testing.T) {
	os.Clearenv()
	os.Setenv("CLEANUP_FREQUENCY_HOURS", "42")
	os.Setenv("CLEANUP_FREQUENCY", "19")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 42
	result := opts.CleanupFrequencyHours()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_FREQUENCY_HOURS value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultCleanupArchiveReadDaysValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 60
	result := opts.CleanupArchiveReadDays()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_ARCHIVE_READ_DAYS value, got %v instead of %v`, result, expected)
	}
}

func TestCleanupArchiveReadDays(t *testing.T) {
	os.Clearenv()
	os.Setenv("CLEANUP_ARCHIVE_READ_DAYS", "7")
	os.Setenv("ARCHIVE_READ_DAYS", "19")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 7
	result := opts.CleanupArchiveReadDays()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_ARCHIVE_READ_DAYS value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultCleanupRemoveSessionsDaysValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 30
	result := opts.CleanupRemoveSessionsDays()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_REMOVE_SESSIONS_DAYS value, got %v instead of %v`, result, expected)
	}
}

func TestCleanupRemoveSessionsDays(t *testing.T) {
	os.Clearenv()
	os.Setenv("CLEANUP_REMOVE_SESSIONS_DAYS", "7")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 7
	result := opts.CleanupRemoveSessionsDays()

	if result != expected {
		t.Fatalf(`Unexpected CLEANUP_REMOVE_SESSIONS_DAYS value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultWorkerPoolSizeValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 42
	result := opts.WorkerPoolSize()

	if result != expected {
		t.Fatalf(`Unexpected WORKER_POOL_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestDefautPollingFrequencyValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 42
	result := opts.PollingFrequency()

	if result != expected {
		t.Fatalf(`Unexpected POLLING_FREQUENCY value, got %v instead of %v`, result, expected)
	}
}

func TestDefaultBatchSizeValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 42
	result := opts.BatchSize()

	if result != expected {
		t.Fatalf(`Unexpected BATCH_SIZE value, got %v instead of %v`, result, expected)
	}
}

func TestDefautPollingSchedulerValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultPollingScheduler
	result := opts.PollingScheduler()

	if result != expected {
		t.Fatalf(`Unexpected POLLING_SCHEDULER value, got %v instead of %v`, result, expected)
	}
}

func TestPollingScheduler(t *testing.T) {
	os.Clearenv()
	os.Setenv("POLLING_SCHEDULER", "entry_count_based")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "entry_count_based"
	result := opts.PollingScheduler()

	if result != expected {
		t.Fatalf(`Unexpected POLLING_SCHEDULER value, got %v instead of %v`, result, expected)
	}
}

func TestDefautSchedulerCountBasedMaxIntervalValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultSchedulerEntryFrequencyMaxInterval
	result := opts.SchedulerEntryFrequencyMaxInterval()

	if result != expected {
		t.Fatalf(`Unexpected SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL value, got %v instead of %v`, result, expected)
	}
}

func TestDefautSchedulerCountBasedMaxInterval(t *testing.T) {
	os.Clearenv()
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL", "30")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 30
	result := opts.SchedulerEntryFrequencyMaxInterval()

	if result != expected {
		t.Fatalf(`Unexpected SCHEDULER_ENTRY_FREQUENCY_MAX_INTERVAL value, got %v instead of %v`, result, expected)
	}
}

func TestDefautSchedulerCountBasedMinIntervalValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultSchedulerEntryFrequencyMinInterval
	result := opts.SchedulerEntryFrequencyMinInterval()

	if result != expected {
		t.Fatalf(`Unexpected SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL value, got %v instead of %v`, result, expected)
	}
}

func TestDefautSchedulerCountBasedMinInterval(t *testing.T) {
	os.Clearenv()
	os.Setenv("SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL", "30")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 30
	result := opts.SchedulerEntryFrequencyMinInterval()

	if result != expected {
		t.Fatalf(`Unexpected SCHEDULER_ENTRY_FREQUENCY_MIN_INTERVAL value, got %v instead of %v`, result, expected)
	}
}

func TestPollingParsingErrorLimit(t *testing.T) {
	os.Clearenv()
	os.Setenv("POLLING_PARSING_ERROR_LIMIT", "100")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 100
	result := opts.PollingParsingErrorLimit()

	if result != expected {
		t.Fatalf(`Unexpected POLLING_SCHEDULER value, got %v instead of %v`, result, expected)
	}
}

func TestOAuth2UserCreationWhenUnset(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "foobar"
	result := opts.OAuth2ClientID()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_ID value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ClientIDValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "secret"
	result := opts.OAuth2ClientSecret()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_CLIENT_SECRET value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ClientSecretValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "http://example.org"
	result := opts.OAuth2RedirectURL()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_REDIRECT_URL value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2RedirectURLValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultOAuth2RedirectURL
	result := opts.OAuth2RedirectURL()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_REDIRECT_URL value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2OidcDiscoveryEndpoint(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_OIDC_DISCOVERY_ENDPOINT", "http://example.org")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "http://example.org"
	result := opts.OAuth2OidcDiscoveryEndpoint()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_OIDC_DISCOVERY_ENDPOINT value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2OidcDiscoveryEndpointValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultOAuth2OidcDiscoveryEndpoint
	result := opts.OAuth2OidcDiscoveryEndpoint()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_REDIRECT_URL value, got %q instead of %q`, result, expected)
	}
}

func TestOAuth2Provider(t *testing.T) {
	os.Clearenv()
	os.Setenv("OAUTH2_PROVIDER", "google")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "google"
	result := opts.OAuth2Provider()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_PROVIDER value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultOAuth2ProviderValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultOAuth2Provider
	result := opts.OAuth2Provider()

	if result != expected {
		t.Fatalf(`Unexpected OAUTH2_PROVIDER value, got %q instead of %q`, result, expected)
	}
}

func TestHSTSWhenUnset(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := false
	result := opts.HasHSTS()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HSTS value, got %v instead of %v`, result, expected)
	}
}

func TestDisableHTTPServiceWhenUnset(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := false
	result := opts.HasHTTPService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_HTTP_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestDisableSchedulerServiceWhenUnset(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := false
	result := opts.HasSchedulerService()

	if result != expected {
		t.Fatalf(`Unexpected DISABLE_SCHEDULER_SERVICE value, got %v instead of %v`, result, expected)
	}
}

func TestRunMigrationsWhenUnset(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := true
	result := opts.RunMigrations()

	if result != expected {
		t.Fatalf(`Unexpected RUN_MIGRATIONS value, got %v instead of %v`, result, expected)
	}
}

func TestCreateAdminWhenUnset(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "something"
	result := opts.PocketConsumerKey("default")

	if result != expected {
		t.Fatalf(`Unexpected POCKET_CONSUMER_KEY value, got %q instead of %q`, result, expected)
	}
}

func TestPocketConsumerKeyFromUserPrefs(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
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

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "all"
	result := opts.ProxyImages()

	if result != expected {
		t.Fatalf(`Unexpected PROXY_IMAGES value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultProxyImagesValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultProxyImages
	result := opts.ProxyImages()

	if result != expected {
		t.Fatalf(`Unexpected PROXY_IMAGES value, got %q instead of %q`, result, expected)
	}
}

func TestHTTPSOff(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	if opts.HTTPS {
		t.Fatalf(`Unexpected HTTPS value, got "%v"`, opts.HTTPS)
	}
}

func TestHTTPSOn(t *testing.T) {
	os.Clearenv()
	os.Setenv("HTTPS", "on")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	if !opts.HTTPS {
		t.Fatalf(`Unexpected HTTPS value, got "%v"`, opts.HTTPS)
	}
}

func TestHTTPClientTimeout(t *testing.T) {
	os.Clearenv()
	os.Setenv("HTTP_CLIENT_TIMEOUT", "42")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := 42
	result := opts.HTTPClientTimeout()

	if result != expected {
		t.Fatalf(`Unexpected HTTP_CLIENT_TIMEOUT value, got %d instead of %d`, result, expected)
	}
}

func TestDefaultHTTPClientTimeoutValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultHTTPClientTimeout
	result := opts.HTTPClientTimeout()

	if result != expected {
		t.Fatalf(`Unexpected HTTP_CLIENT_TIMEOUT value, got %d instead of %d`, result, expected)
	}
}

func TestHTTPClientMaxBodySize(t *testing.T) {
	os.Clearenv()
	os.Setenv("HTTP_CLIENT_MAX_BODY_SIZE", "42")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := int64(42 * 1024 * 1024)
	result := opts.HTTPClientMaxBodySize()

	if result != expected {
		t.Fatalf(`Unexpected HTTP_CLIENT_MAX_BODY_SIZE value, got %d instead of %d`, result, expected)
	}
}

func TestDefaultHTTPClientMaxBodySizeValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := int64(defaultHTTPClientMaxBodySize * 1024 * 1024)
	result := opts.HTTPClientMaxBodySize()

	if result != expected {
		t.Fatalf(`Unexpected HTTP_CLIENT_MAX_BODY_SIZE value, got %d instead of %d`, result, expected)
	}
}

func TestParseConfigFile(t *testing.T) {
	content := []byte(`
 # This is a comment

DEBUG = yes

 POCKET_CONSUMER_KEY= >#1234

Invalid text
`)

	tmpfile, err := os.CreateTemp(".", "miniflux.*.unit_test.conf")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}

	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseFile(tmpfile.Name())
	if err != nil {
		t.Errorf(`Parsing failure: %v`, err)
	}

	if opts.HasDebugMode() != true {
		t.Errorf(`Unexpected debug mode value, got "%v"`, opts.HasDebugMode())
	}

	expected := ">#1234"
	result := opts.PocketConsumerKey("default")
	if result != expected {
		t.Errorf(`Unexpected POCKET_CONSUMER_KEY value, got %q instead of %q`, result, expected)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(tmpfile.Name()); err != nil {
		t.Fatal(err)
	}
}

func TestAuthProxyHeader(t *testing.T) {
	os.Clearenv()
	os.Setenv("AUTH_PROXY_HEADER", "X-Forwarded-User")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := "X-Forwarded-User"
	result := opts.AuthProxyHeader()

	if result != expected {
		t.Fatalf(`Unexpected AUTH_PROXY_HEADER value, got %q instead of %q`, result, expected)
	}
}

func TestDefaultAuthProxyHeaderValue(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := defaultAuthProxyHeader
	result := opts.AuthProxyHeader()

	if result != expected {
		t.Fatalf(`Unexpected AUTH_PROXY_HEADER value, got %q instead of %q`, result, expected)
	}
}

func TestAuthProxyUserCreationWhenUnset(t *testing.T) {
	os.Clearenv()

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := false
	result := opts.IsAuthProxyUserCreationAllowed()

	if result != expected {
		t.Fatalf(`Unexpected AUTH_PROXY_USER_CREATION value, got %v instead of %v`, result, expected)
	}
}

func TestAuthProxyUserCreationAdmin(t *testing.T) {
	os.Clearenv()
	os.Setenv("AUTH_PROXY_USER_CREATION", "1")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := true
	result := opts.IsAuthProxyUserCreationAllowed()

	if result != expected {
		t.Fatalf(`Unexpected AUTH_PROXY_USER_CREATION value, got %v instead of %v`, result, expected)
	}
}

func TestFetchYouTubeWatchTime(t *testing.T) {
	os.Clearenv()
	os.Setenv("FETCH_YOUTUBE_WATCH_TIME", "1")

	parser := NewParser()
	opts, err := parser.ParseEnvironmentVariables()
	if err != nil {
		t.Fatalf(`Parsing failure: %v`, err)
	}

	expected := true
	result := opts.FetchYouTubeWatchTime()

	if result != expected {
		t.Fatalf(`Unexpected FETCH_YOUTUBE_WATCH_TIME value, got %v instead of %v`, result, expected)
	}
}

func TestParseConfigDumpOutput(t *testing.T) {
	os.Clearenv()

	wantOpts := NewOptions()
	wantOpts.adminUsername = "my-username"

	serialized := wantOpts.String()
	tmpfile, err := os.CreateTemp(".", "miniflux.*.unit_test.conf")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte(serialized)); err != nil {
		t.Fatal(err)
	}

	parser := NewParser()
	parsedOpts, err := parser.ParseFile(tmpfile.Name())
	if err != nil {
		t.Errorf(`Parsing failure: %v`, err)
	}

	if parsedOpts.AdminUsername() != wantOpts.AdminUsername() {
		t.Fatalf(`Unexpected ADMIN_USERNAME value, got %q instead of %q`, parsedOpts.AdminUsername(), wantOpts.AdminUsername())
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(tmpfile.Name()); err != nil {
		t.Fatal(err)
	}
}
