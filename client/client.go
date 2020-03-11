// client.go - HTTP client related functionality.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package client provides HTTP client related syntactic sugar.
//
// * GET call with context
//
//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//   defer cancel()
//
//  resp := client.MustGetWithContext(ctx, "https://domain.tld")
//   defer resp.Body.Close()
//
// * POST call with context
//
//   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//   defer cancel()
//
//   resp := client.MustPostWithContext(ctx, "https://domain.tld", "application/json", strings.NewReader("{}"))
//   defer resp.Body.Close()
package client

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// httpClientDo is a shortcut to GetHttpClient().Do. It helps write testable code.
var httpClientDo = GetHttpClient().Do

// httpNewRequestWithContext is a shortcut to http.NewRequestWithContext.
// It helps write testable code.
var httpNewRequestWithContext = http.NewRequestWithContext

// logFatalf is a shortcut to log.Fatalf. It helps write testable code.
var logFatalf = log.Fatalf

// GetHttpClient provides http client with configured connection and timeout.
func GetHttpClient() *http.Client {
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

// MustGetWithContext is a helper that wraps a http GET call to given URL and
// log any error.
func MustGetWithContext(ctx context.Context, url string) *http.Response {
	log.Printf("GET %s", url)

	req, err := httpNewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logFatalf("GET %s failed: %s", url, err)
	}

	resp, err := httpClientDo(req)
	if err != nil {
		logFatalf("GET %s failed: %s", url, err)
	}

	return resp
}

// MustPostWithContext is a helper that wraps a http POST call to given URL,
// content type, and body, as well as log any error.
func MustPostWithContext(ctx context.Context, url, contentType string, body *strings.Reader) *http.Response {
	log.Printf("POST %s %+v", url, body)

	req, err := httpNewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		logFatalf("POST %s failed: %s", url, err)
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := httpClientDo(req)
	if err != nil {
		logFatalf("POST %s failed: %s", url, err)
	}

	return resp
}
