// packer.go - Unpack related functionality.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package packer provides extracting archive related syntactic sugar.
//
// * Extract tar.gz content
//
//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//   defer cancel()
//
//   resp := client.MustGetWithContext(ctx, "https://domain.tld")
//   defer resp.Body.Close()
//
//   packer.Untar(resp.Body, "/path/to/extract")
//
// * Extract zip content
//
//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//   defer cancel()
//
//   resp := client.MustGetWithContext(ctx, "https://domain.tld")
//   defer resp.Body.Close()
//
//   packer.Unzip(resp.Body, "/path/to/extract")
package packer

import (
	"log"
	"os"
)

// logFatalf is a shortcut to log.Fatalf. It helps write testable code.
var logFatalf = log.Fatalf

// osRemove is a shortcut to os.Remove. It helps write testable code.
var osRemove = os.Remove
