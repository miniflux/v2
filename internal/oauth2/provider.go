// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package oauth2 // import "miniflux.app/v2/internal/oauth2"

import (
	"context"

	"golang.org/x/oauth2"

	"miniflux.app/v2/internal/model"
)

// Provider defines the interface that all OAuth2 providers must implement.
type Provider interface {
	// Config returns the OAuth2 configuration for this provider.
	Config() *oauth2.Config

	// UserExtraKey returns the key used to store the provider-specific user ID.
	UserExtraKey() string

	// Profile exchanges the authorization code for a token and fetches the user's profile.
	Profile(ctx context.Context, code, codeVerifier string) (*UserProfile, error)

	// PopulateUserCreationWithProfileID sets the provider-specific ID on a new user creation request.
	PopulateUserCreationWithProfileID(user *model.UserCreationRequest, profile *UserProfile)

	// PopulateUserWithProfileID sets the provider-specific ID on an existing user.
	PopulateUserWithProfileID(user *model.User, profile *UserProfile)

	// UserProfileID returns the provider-specific ID from the given user.
	UserProfileID(user *model.User) string

	// UnsetUserProfileID removes the provider-specific ID from the given user.
	UnsetUserProfileID(user *model.User)
}
