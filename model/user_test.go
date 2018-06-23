// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import "testing"

func TestValidateUserCreation(t *testing.T) {
	user := &User{}
	if err := user.ValidateUserCreation(); err == nil {
		t.Error(`An empty user should generate an error`)
	}

	user = &User{Username: "test", Password: ""}
	if err := user.ValidateUserCreation(); err == nil {
		t.Error(`User without password should generate an error`)
	}

	user = &User{Username: "test", Password: "a"}
	if err := user.ValidateUserCreation(); err == nil {
		t.Error(`Passwords shorter than 6 characters should generate an error`)
	}

	user = &User{Username: "", Password: "secret"}
	if err := user.ValidateUserCreation(); err == nil {
		t.Error(`An empty username should generate an error`)
	}

	user = &User{Username: "test", Password: "secret"}
	if err := user.ValidateUserCreation(); err != nil {
		t.Error(`A valid user should not generate any error`)
	}
}

func TestValidateUserModification(t *testing.T) {
	user := &User{}
	if err := user.ValidateUserModification(); err != nil {
		t.Error(`There is no changes, so we should not have an error`)
	}

	user = &User{Theme: "default"}
	if err := user.ValidateUserModification(); err != nil {
		t.Error(`A valid theme should not generate any errors`)
	}

	user = &User{Theme: "invalid theme"}
	if err := user.ValidateUserModification(); err == nil {
		t.Error(`An invalid theme should generate an error`)
	}

	user = &User{Password: "test123"}
	if err := user.ValidateUserModification(); err != nil {
		t.Error(`A valid password should not generate any errors`)
	}

	user = &User{Password: "a"}
	if err := user.ValidateUserModification(); err == nil {
		t.Error(`An invalid password should generate an error`)
	}
}
