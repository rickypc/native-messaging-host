// download.go - Fetch updates.xml and download latest file content.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package host

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/rickypc/native-messaging-host/client"
	"io"
	"net/http"
	"os"
	"time"
)

// fs is a shortcut to *FileSystem. It helps write testable code.
var fs FileSystemInterface = &FileSystem{}

// ioCopy is a shortcut to io.Copy. It helps write testable code.
var ioCopy = io.Copy

// osRename is a shortcut to os.Rename. It helps write testable code.
var osRename = os.Rename

// FileInterface is an interface for OpenFile first-value return. It helps write
// testable code.
type FileInterface interface {
	io.Closer
	io.Writer
}

// FileSystemInterface is an interface for OpenFile to be overridable. It helps
// write testable code.
type FileSystemInterface interface {
	OpenFile(name string, flag int, perm os.FileMode) (FileInterface, error)
}

// FileSystem is an implementation of FileSystemInterface. It helps write
// testable code.
type FileSystem struct{}

// OpenFile is an implementation of FileSystemInterface.OpenFile and wrap
// os.OpenFile. It helps write testable code.
func (f *FileSystem) OpenFile(name string, flag int, perm os.FileMode) (FileInterface, error) {
	return os.OpenFile(name, flag, perm)
}

// downloadLatest will download latest file content from given download URL and
// replace current executable with it. It will return error when it come across
// one.
func (h *Host) downloadLatest(url string) error {
	ctx, cancel := context.WithTimeout(context.Background(), HttpOverallTimeout*time.Second)
	defer cancel()

	resp := client.MustGetWithContext(ctx, url)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to find the update: %d", resp.StatusCode)
	}

	backupName := h.ExecName + ".bak"
	if err := osRename(h.ExecName, backupName); err != nil {
		return err
	}

	file, err := fs.OpenFile(h.ExecName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		if mvErr := osRename(backupName, h.ExecName); mvErr != nil {
			err = fmt.Errorf("%w %v", err, mvErr)
		}
		return err
	}
	defer file.Close()

	if _, err := ioCopy(file, resp.Body); err != nil {
		if mvErr := osRename(backupName, h.ExecName); mvErr != nil {
			err = fmt.Errorf("%w %v", err, mvErr)
		}
		return err
	}

	os.Remove(backupName)
	return nil
}

// getDownloadUrlAndVersion returns download URL and latest version on
// configured application name. It will return error when it come across one.
func (h *Host) getDownloadUrlAndVersion() (string, string, error) {
	url := ""
	version := ""

	ctx, cancel := context.WithTimeout(context.Background(), HttpOverallTimeout*time.Second)
	defer cancel()

	resp := client.MustGetWithContext(ctx, h.UpdateUrl)
	defer resp.Body.Close()

	response := &UpdateCheckResponse{}
	if err := xml.NewDecoder(resp.Body).Decode(response); err != nil {
		return url, version, err
	}

	url, version = response.GetUrlAndVersion(h.AppName)
	return url, version, nil
}
