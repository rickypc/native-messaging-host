// manifest.go - Install and Uninstall manifest file for Linux.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// +build !darwin,!windows

package host

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

// getTargetName returns an absolute path to native messaging host manifest
// location for Linux.
// See https://developer.chrome.com/extensions/nativeMessaging#native-messaging-host-location-nix
func (h *Host) getTargetName() (string, error) {
	target := "/etc/opt/chrome/native-messaging-hosts"

	current, err := user.Current()
	if err != nil {
		return "", err
	}

	if current.Uid != "0" {
		target = current.HomeDir + "/.config/google-chrome/NativeMessagingHosts"
	}

	return filepath.Join(target, h.AppName+".json"), nil
}

// Install creates native-messaging manifest file on appropriate location.
func (h *Host) Install() error {
	targetName, err := h.getTargetName()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(targetName), 0755); err != nil {
		return err
	}

	manifest, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(targetName, manifest, 0644); err != nil {
		return err
	}

	log.Printf("Installed: %s", targetName)
	return nil
}

// Uninstall removes native-messaging manifest file from installed location.
func (h *Host) Uninstall() error {
	targetName, err := h.getTargetName()
	if err != nil {
		return err
	}

	if err := os.Remove(targetName); err != nil {
		// It might never have been installed.
		log.Print(err)
	}

	if err := os.Remove(h.ExecName); err != nil {
		// It might be locked by current process.
		log.Print(err)
	}

	if err := os.Remove(h.ExecName + ".chk"); err != nil {
		// It might not exist.
		log.Print(err)
	}

	log.Printf("Uninstalled: %s", targetName)

	// Exit gracefully.
	runtime.Goexit()
	return nil
}
