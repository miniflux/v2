// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package model

import "testing"

func TestValidateCategoryCreation(t *testing.T) {
	category := &Category{}
	if err := category.ValidateCategoryCreation(); err == nil {
		t.Error(`An empty category should generate an error`)
	}

	category = &Category{Title: "Test"}
	if err := category.ValidateCategoryCreation(); err == nil {
		t.Error(`A category without userID should generate an error`)
	}

	category = &Category{UserID: 42}
	if err := category.ValidateCategoryCreation(); err == nil {
		t.Error(`A category without title should generate an error`)
	}

	category = &Category{Title: "Test", UserID: 42}
	if err := category.ValidateCategoryCreation(); err != nil {
		t.Error(`All required fields are filled, it should not generate any error`)
	}
}

func TestValidateCategoryModification(t *testing.T) {
	category := &Category{}
	if err := category.ValidateCategoryModification(); err == nil {
		t.Error(`An empty category should generate an error`)
	}

	category = &Category{Title: "Test"}
	if err := category.ValidateCategoryModification(); err == nil {
		t.Error(`A category without userID should generate an error`)
	}

	category = &Category{UserID: 42}
	if err := category.ValidateCategoryModification(); err == nil {
		t.Error(`A category without title should generate an error`)
	}

	category = &Category{ID: -1, Title: "Test", UserID: 42}
	if err := category.ValidateCategoryModification(); err == nil {
		t.Error(`An invalid categoryID should generate an error`)
	}

	category = &Category{ID: 0, Title: "Test", UserID: 42}
	if err := category.ValidateCategoryModification(); err == nil {
		t.Error(`An invalid categoryID should generate an error`)
	}

	category = &Category{ID: 1, Title: "Test", UserID: 42}
	if err := category.ValidateCategoryModification(); err != nil {
		t.Error(`All required fields are filled, it should not generate any error`)
	}
}
