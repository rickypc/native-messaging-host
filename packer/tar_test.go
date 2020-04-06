// tar_test.go - Test for tar related functionality.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package packer

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestTarRemoveLink(t *testing.T) {
	t.Parallel()

	compare := func(wantErr int) func(t *testing.T) {
		return func(t *testing.T) {
			fatal := false
			oldLogFatalf := logFatalf
			oldOsRemove := osRemove
			removed := 0
			targetName := fmt.Sprintf("../testdata/tarlink-%d", wantErr)

			defer func() {
				_ = recover()
				logFatalf = oldLogFatalf
				osRemove = oldOsRemove
			}()

			logFatalf = func(msg string, v ...interface{}) {
				fatal = true
				panic(fmt.Sprintf(msg, v))
			}
			osRemove = func(string) error { removed++; return nil }

			switch wantErr {
			case 0:
				if err := ioutil.WriteFile(targetName, []byte(""), 0644); err != nil {
					t.Fatalf("touch file error: %v", err)
				}
				defer func() { os.Remove(targetName) }()
			case 2:
				if err := ioutil.WriteFile(targetName, []byte(""), 0644); err != nil {
					t.Fatalf("touch file error: %v", err)
				}
				defer func() { os.Remove(targetName) }()
				osRemove = func(string) error {
					removed++
					return errors.New("remove error")
				}
			}

			removeLink(targetName)

			switch wantErr {
			case 0:
				if fatal || removed < 1 {
					t.Errorf("should not panic and removed: %v, %d", fatal, removed)
				}
			case 1:
				if fatal || removed > 0 {
					t.Errorf("should not panic and not removed: %v, %d", fatal, removed)
				}
			case 2:
				if !fatal || removed > 0 {
					t.Errorf("should panic and not removed: %v, %d", fatal, removed)
				}
			}
		}
	}

	t.Run("with file exists", compare(0))
	t.Run("with file non-exist", compare(1))
	t.Run("with remove error", compare(2))
}

func TestTarUntar(t *testing.T) {
	t.Parallel()

	compare := func(wantErr int) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			target := "../testdata/untar"
			file, _ := os.Open("../testdata/packer.tgz")
			Untar(file, target)

			switch wantErr {
			case 0:
			}

			os.RemoveAll(target)
		}
	}

	t.Run("with valid file", compare(0))
}

func TestTarValidRelPath(t *testing.T) {
	t.Parallel()

	compare := func(want bool, p string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			if got := validRelPath(p); got != want {
				t.Errorf("mismatch %s (want: %t, got: %t)", p, want, got)
			}
		}
	}

	t.Run("with relative path", compare(true, "./path/to/nowhere"))
	t.Run("with absolute path", compare(false, "/path/to/nowhere"))
	t.Run("with nothing", compare(false, ""))
	t.Run(`with "\"`, compare(false, `path\to\nowhere`))
	t.Run(`with "../"`, compare(false, "../path/to/nowhere"))
}
