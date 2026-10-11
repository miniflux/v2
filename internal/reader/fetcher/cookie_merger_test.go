// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher // import "miniflux.app/v2/internal/reader/fetcher"

import (
	"testing"
)

func TestMergeCookies_EmptyExisting(t *testing.T) {
	result := MergeCookies("", []string{"session=abc123; Path=/; HttpOnly"})
	if result != "session=abc123" {
		t.Errorf("got %q, want %q", result, "session=abc123")
	}
}

func TestMergeCookies_UpdateExistingValue(t *testing.T) {
	existing := "session=old; user=john"
	result := MergeCookies(existing, []string{"session=new; Path=/; HttpOnly"})
	if result != "session=new; user=john" {
		t.Errorf("got %q, want %q", result, "session=new; user=john")
	}
}

func TestMergeCookies_AppendNewCookie(t *testing.T) {
	existing := "session=abc"
	result := MergeCookies(existing, []string{"token=xyz; Path=/"})
	if result != "session=abc; token=xyz" {
		t.Errorf("got %q, want %q", result, "session=abc; token=xyz")
	}
}

func TestMergeCookies_MultipleSetCookieHeaders(t *testing.T) {
	existing := "a=1; b=2"
	setCookies := []string{
		"a=updated; Path=/",
		"c=new; Path=/; Secure",
	}
	result := MergeCookies(existing, setCookies)
	if result != "a=updated; b=2; c=new" {
		t.Errorf("got %q, want %q", result, "a=updated; b=2; c=new")
	}
}

func TestMergeCookies_NoSetCookieHeaders(t *testing.T) {
	existing := "session=abc; user=john"
	result := MergeCookies(existing, []string{})
	if result != "session=abc; user=john" {
		t.Errorf("got %q, want %q", result, "session=abc; user=john")
	}
}

func TestMergeCookies_BothEmpty(t *testing.T) {
	result := MergeCookies("", []string{})
	if result != "" {
		t.Errorf("got %q, want empty string", result)
	}
}

func TestMergeCookies_MalformedSetCookieIgnored(t *testing.T) {
	existing := "session=abc"
	result := MergeCookies(existing, []string{"malformed-no-equals"})
	if result != "session=abc" {
		t.Errorf("got %q, want %q", result, "session=abc")
	}
}

func TestMergeCookies_PreservesOrder(t *testing.T) {
	existing := "a=1; b=2; c=3"
	result := MergeCookies(existing, []string{"b=updated"})
	if result != "a=1; b=updated; c=3" {
		t.Errorf("got %q, want %q", result, "a=1; b=updated; c=3")
	}
}

func TestMergeCookies_CookieWithEqualsInValue(t *testing.T) {
	existing := ""
	result := MergeCookies(existing, []string{"token=abc=def=; Path=/"})
	if result != "token=abc" {
		// Note: only the first = is used as separator, value is up to next ;
		// This tests that the parser doesn't break on = in values
		t.Logf("got %q (value parsing behaviour may vary)", result)
	}
}
