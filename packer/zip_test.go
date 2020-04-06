// zip_test.go - Test for zip related functionality.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package packer

import (
	"os"
	"testing"
)

func TestZipUnzip(t *testing.T) {
	t.Parallel()

	compare := func(wantErr int) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			target := "../testdata/unzip"
			file, _ := os.Open("../testdata/packer.zip")
			Unzip(file, target)

			switch wantErr {
			case 0:
			}

			os.RemoveAll(target)
		}
	}

	t.Run("with valid file", compare(0))
}
