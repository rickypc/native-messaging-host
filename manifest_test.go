// manifest_test.go - Test for manifest related functionality on Linux.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build !darwin,!windows

package host

import (
	"encoding/json"
	"errors"
	"github.com/google/go-cmp/cmp"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestManifestTargetName(t *testing.T) {
	t.Parallel()

	got := (&Host{AppName: "app"}).getTargetName()
	homeDir, _ := os.UserHomeDir()
	want := homeDir + "/.config/google-chrome/NativeMessagingHosts/app.json"

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestManifestInstall(t *testing.T) {
	t.Parallel()

	log.SetOutput(ioutil.Discard)

	compare := func(wantErr int, uninstall bool) func(t *testing.T) {
		return func(t *testing.T) {
			got := &Host{}
			want := &Host{AppName: "install"}
			targetName := want.getTargetName()

			switch wantErr {
			case 0:
				if uninstall {
					os.Remove(targetName)
				}
			case 1:
				oldOsMkdirAll := osMkdirAll
				defer func() { osMkdirAll = oldOsMkdirAll }()
				osMkdirAll = func(string, os.FileMode) error {
					return errors.New("MkdirAll error")
				}
			case 2:
				oldWriteFile := ioutilWriteFile
				defer func() { ioutilWriteFile = oldWriteFile }()
				ioutilWriteFile = func(string, []byte, os.FileMode) error {
					return errors.New("WriteFile error")
				}
			}

			if err := want.Install(); wantErr == 0 && err != nil {
				t.Errorf("install error %s: %v", targetName, err)
			} else if wantErr > 0 && err == nil {
				t.Fatalf("want error: %s", targetName)
			}

			if wantErr == 0 {
				if _, err := os.Stat(targetName); err != nil {
					t.Errorf("missing file %s: %v", targetName, err)
				}

				manifest, err := ioutil.ReadFile(targetName)
				if err != nil {
					t.Errorf("read manifest error %s: %v", targetName, err)
				}

				if err := json.Unmarshal(manifest, got); err != nil {
					t.Errorf("unmarshal manifest error %s: %v", targetName, err)
				}

				if diff := cmp.Diff(want, got); diff != "" {
					t.Errorf("mismatch (-want +got):\n%s", diff)
				}
			}
		}
	}

	t.Run("with nothing installed", compare(0, false))
	t.Run("with existing installed", compare(0, true))
	t.Run("with MkdirAll error", compare(1, false))
	t.Run("with WriteFile error", compare(2, false))
}

func TestManifestUninstall(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	compare := func(h *Host) func(t *testing.T) {
		return func(t *testing.T) {
			exited := false
			oldRuntimeGoexit := runtimeGoexit
			defer func() { runtimeGoexit = oldRuntimeGoexit }()
			runtimeGoexit = func() { exited = true }
			targetName := h.getTargetName()

			h.Uninstall()

			if _, err := os.Stat(targetName); err == nil {
				t.Errorf("uninstall failed %s", targetName)
			}

			if !exited {
				t.Errorf("uninstall did not exit")
			}
		}
	}

	h := &Host{AppName: "uninstall"}

	t.Run("with nothing installed", compare(h))

	if err := h.Install(); err != nil {
		t.Errorf("install error %s: %v", h.getTargetName(), err)
	}

	t.Run("with installed", compare(h))
}
