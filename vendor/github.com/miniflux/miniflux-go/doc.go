// Copyright 2017 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

/*

Package miniflux implements a client library for the Miniflux REST API.

Examples

This code snippet fetch the list of users.

	client := miniflux.NewClient("https://api.example.org", "admin", "secret")
	users, err := client.Users()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(users, err)

This one discover subscriptions on a website.

	subscriptions, err := client.Discover("https://example.org/")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(subscriptions)

*/
package miniflux
