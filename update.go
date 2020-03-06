// update.go - Reads and find latest update from updates.xml.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package host

import (
	"github.com/hashicorp/go-version"
	"io/ioutil"
	"log"
	"strconv"
	"time"
)

// AutoUpdateCheck downloads the latest update as necessary.
func (h *Host) AutoUpdateCheck() {
	if h.AutoUpdate {
		if needed, downloadUrl := h.needUpdate(); needed {
			if err := h.downloadLatest(downloadUrl); err != nil {
				log.Printf("Update download error: %v", err)
			} else {
				log.Print("Update is downloaded")
			}
		}
	}
}

// getCheckTimestamp returns previous update check timestamp in Unix
// nanoseconds.
func (h *Host) getCheckTimestamp() time.Time {
	buf, _ := ioutil.ReadFile(h.ExecName + ".chk")
	nano, _ := strconv.ParseInt(string(buf), 10, 64)
	return time.Unix(0, nano)
}

// isCheckedToday returns true if update check was done sometime today,
// otherwise false.
func (h *Host) isCheckedToday() bool {
	y1, m1, d1 := h.getCheckTimestamp().Date()
	y2, m2, d2 := time.Now().Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// needUpdate returns true if update is needed, otherwise false.
//
// Truthy criteria:
// - Update check wasn't already done sometime today.
// - Current running version is older than updates.xml's version.
func (h *Host) needUpdate() (bool, string) {
	response := false

	if h.isCheckedToday() {
		log.Print("Update already checked today")
		return response, ""
	}

	if err := h.writeCheckTimestamp(); err != nil {
		log.Printf("Update timestamp error: %v", err)
	}

	localVersion := version.Must(version.NewVersion(h.Version))

	downloadUrl, remoteRawVersion, err := h.getDownloadUrlAndVersion()
	if err != nil {
		log.Printf("Update check error: %v", err)
	}

	remoteVersion := version.Must(version.NewVersion(remoteRawVersion))

	if localVersion.LessThan(remoteVersion) {
		log.Print("Latest update is found")
		response = true
	} else {
		log.Print("Already up to date")
	}

	return response, downloadUrl
}

// writeCheckTimestamp writes update check timestamp in Unix nanoseconds.
// It will return error when it unable to write to .chk file.
func (h *Host) writeCheckTimestamp() error {
	timestamp := []byte(strconv.FormatInt(time.Now().UnixNano(), 10))

	if err := ioutil.WriteFile(h.ExecName+".chk", timestamp, 0644); err != nil {
		return err
	}

	return nil
}
