// Copyright 2021 Frédéric Guillot. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package cli // import "miniflux.app/cli"

import (
	"net"
	"os"
)

const (
	// sdNotifyReady tells the service manager that service startup is
	// finished, or the service finished loading its configuration.
	sdNotifyReady = "READY=1"
)

// sdNotify sends a message to systemd using the sd_notify protocol.
// See https://www.freedesktop.org/software/systemd/man/sd_notify.html.
func sdNotify(state string) error {
	addr := &net.UnixAddr{
		Net:  "unixgram",
		Name: os.Getenv("NOTIFY_SOCKET"),
	}

	if addr.Name == "" {
		// We're not running under systemd (NOTIFY_SOCKET has not set).
		return nil
	}

	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err = conn.Write([]byte(state)); err != nil {
		return err
	}

	return nil
}
