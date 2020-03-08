// manifest_windows.go - Install and Uninstall manifest file for Windows.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package host

import (
	"encoding/json"
	"golang.org/x/sys/windows/registry"
	"log"
	"os"
	"path/filepath"
)

// Install creates native-messaging manifest file on appropriate location and
// add an entry in windows registry. It will return error when it come across
// one.
//
// See https://developer.chrome.com/extensions/nativeMessaging#native-messaging-host-location
func (h *Host) Install() error {
	manifest, _ := json.MarshalIndent(h, "", "  ")
	registryName := `Software\Google\Chrome\NativeMessagingHosts\` + h.AppName
	targetName := filepath.Join(filepath.Dir(h.ExecName), h.AppName+".json")

	if err := ioutilWriteFile(targetName, manifest, 0644); err != nil {
		return err
	}

	key, err := registry.OpenKey(registry.CURRENT_USER, registryName, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if err := key.SetStringValue("", targetName); err != nil {
		return err
	}

	log.Printf(`Installed: HKCU\%s`, registryName)
	return nil
}

// Uninstall removes entry from windows registry and removes native-messaging
// manifest file from installed location.
//
// See https://developer.chrome.com/extensions/nativeMessaging#native-messaging-host-location
func (h *Host) Uninstall() {
	registryName := `Software\Google\Chrome\NativeMessagingHosts\` + h.AppName
	targetName := filepath.Join(filepath.Dir(h.ExecName), h.AppName+".json")

	key, err := registry.OpenKey(registry.CURRENT_USER, registryName, registry.SET_VALUE)
	if err != nil {
		// Unable to open windows registry.
		log.Print(err)
	}
	defer key.Close()

	if err := key.DeleteValue(""); err != nil {
		// It might never have been installed.
		log.Print(err)
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

	log.Printf(`Uninstalled: HKCU\%s`, registryName)

	// Exit gracefully.
	runtimeGoexit()
}
