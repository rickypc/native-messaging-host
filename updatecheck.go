// updatecheck.go - Reads and find latest update from updates.xml.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package host

import (
	"encoding/xml"
	"runtime"
)

// An App is represent one application returned by updates.xml.
type App struct {
	AppId   *string   `xml:"appid,attr"`
	Updates []*Update `xml:"updatecheck"`
}

// An Update is represent application download URL and latest version.
//
// It can have target OS optionally. This is an extended attribute that is not
// part of original Google Chrome update manifest.
type Update struct {
	Goos    *string `xml:"os,attr"`
	Url     *string `xml:"codebase,attr"`
	Version *string `xml:"version,attr"`
}

// An UpdateCheckResponse implements Google Chrome update manifest XML format
// borrowed from Google's Omaha.
// See https://developer.chrome.com/apps/autoupdate#update_manifest
type UpdateCheckResponse struct {
	Apps    []*App   `xml:"app"`
	XMLName xml.Name `xml:"gupdate"`
}

// getAppId returns application identifier.
func (a *App) getAppId() string {
	if a.AppId != nil {
		return *a.AppId
	}
	return ""
}

// getUrlAndVersion returns application download URL and latest version that
// match runtime.GOOS, otherwise it will return the first available one.
func (a *App) getUrlAndVersion() (string, string) {
	url := ""
	version := ""

	for _, update := range a.Updates {
		if update.getGoos() == runtime.GOOS {
			url = update.getUrl()
			version = update.getVersion()
			break
		}
	}

	if (url == "" || version == "") && len(a.Updates) > 0 {
		update := a.Updates[0]
		url = update.getUrl()
		version = update.getVersion()
	}

	return url, version
}

// getGoos returns application target OS.
func (u *Update) getGoos() string {
	if u.Goos != nil {
		return *u.Goos
	}
	return ""
}

// getUrl returns application download URL.
func (u *Update) getUrl() string {
	if u.Url != nil {
		return *u.Url
	}
	return ""
}

// getVersion returns application latest version.
func (u *Update) getVersion() string {
	if u.Version != nil {
		return *u.Version
	}
	return ""
}

// GetUrlAndVersion returns download URL and latest version of given
// application name.
func (u *UpdateCheckResponse) GetUrlAndVersion(appName string) (string, string) {
	url := ""
	version := ""

	for _, app := range u.Apps {
		if app.getAppId() == appName {
			url, version = app.getUrlAndVersion()
			break
		}
	}

	return url, version
}
