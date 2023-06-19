// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

/*
Package client implements a client library for the Miniflux REST API.

# Examples

This code snippet fetch the list of users:

	import (
		miniflux "miniflux.app/client"
	)

	client := miniflux.New("https://api.example.org", "admin", "secret")
	users, err := client.Users()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(users, err)

This one discover subscriptions on a website:

	subscriptions, err := client.Discover("https://example.org/")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(subscriptions)
*/
package client // import "miniflux.app/client"
