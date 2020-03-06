// transport.go - Fetch updates.xml and download latest file content.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package host

import (
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

// downloadLatest will download latest file content from given download URL and
// replace current executable with it. It will return error when it come across
// one.
func (h *Host) downloadLatest(url string) error {
	log.Printf("GET %s", url)
	resp, err := h.GetHttpClient().Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unable to find the update: %d", resp.StatusCode)
	}

	backupName := h.ExecName + ".bak"
	if err := os.Rename(h.ExecName, backupName); err != nil {
		return err
	}

	file, err := os.OpenFile(h.ExecName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		if mvErr := os.Rename(backupName, h.ExecName); mvErr != nil {
			err = fmt.Errorf("%w %v", err, mvErr)
		}
		return err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		if mvErr := os.Rename(backupName, h.ExecName); mvErr != nil {
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

	if resp := h.MustGet(h.UpdateUrl); resp != nil {
		defer resp.Body.Close()

		response := &UpdateCheckResponse{}
		if err := xml.NewDecoder(resp.Body).Decode(response); err != nil {
			return url, version, err
		}

		url, version = response.GetUrlAndVersion(h.AppName)
	}

	return url, version, nil
}

// GetHttpClient provides http client with configured connection and timeout.
func (h *Host) GetHttpClient() *http.Client {
	httpTransport := &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: HttpKeepAlive * time.Second,
			Timeout:   HttpDialTimeout * time.Second,
		}).DialContext,
		ExpectContinueTimeout: HttpContinueTimeout * time.Second,
		IdleConnTimeout:       IdleTimeout * time.Second,
		MaxIdleConns:          MaxConnections,
		MaxIdleConnsPerHost:   MaxConnections,
		ResponseHeaderTimeout: ResponseHeaderTimeout * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSHandshakeTimeout: TLSDialTimeout * time.Second,
	}

	return &http.Client{
		Timeout:   HttpOverallTimeout * time.Second,
		Transport: httpTransport,
	}
}

// MustGet is a helper that wraps a http GET call to given URL and log error if any.
func (h *Host) MustGet(url string) *http.Response {
	log.Printf("GET %s", url)
	resp, err := h.GetHttpClient().Get(url)

	if err != nil {
		log.Printf("GET %s failed: %s", url, err)
	}

	return resp
}
