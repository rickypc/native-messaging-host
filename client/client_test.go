// client_test.go - Test for HTTP client related functionality.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientMustGetWithContext(t *testing.T) {
	compare := func(wantErr int) func(t *testing.T) {
		return func(t *testing.T) {
			did := false
			fatal := false
			requested := false
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte("OK"))
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			switch wantErr {
			case 1:
				oldLogFatalf := logFatalf
				oldRequest := httpNewRequestWithContext
				defer func() {
					_ = recover()
					httpNewRequestWithContext = oldRequest
					logFatalf = oldLogFatalf
				}()
				httpNewRequestWithContext = func(context.Context, string, string, io.Reader) (*http.Request, error) {
					requested = true
					return nil, errors.New("request error")
				}
				logFatalf = func(msg string, v ...interface{}) {
					fatal = true
					panic(fmt.Sprintf(msg, v))
				}
			case 2:
				oldHttpClientDo := httpClientDo
				oldLogFatalf := logFatalf
				defer func() {
					_ = recover()
					httpClientDo = oldHttpClientDo
					logFatalf = oldLogFatalf
				}()
				httpClientDo = func(*http.Request) (*http.Response, error) {
					did = true
					return nil, errors.New("client error")
				}
				logFatalf = func(msg string, v ...interface{}) {
					fatal = true
					panic(fmt.Sprintf(msg, v))
				}
			}

			resp := MustGetWithContext(ctx, server.URL)

			switch wantErr {
			case 0:
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				if string(body) != "OK" {
					t.Errorf("content mismatch: %s", body)
				}
			case 1:
				if !requested || did || !fatal {
					t.Errorf("wrong journey: %v, %v, %v", requested, did, fatal)
				}
			case 2:
				if !requested || !did || !fatal {
					t.Errorf("wrong journey: %v, %v, %v", requested, did, fatal)
				}
			}
		}
	}

	t.Run("with valid response", compare(0))
	t.Run("with request error", compare(1))
	t.Run("with client error", compare(2))
}

func TestClientMustPostWithContext(t *testing.T) {
	compare := func(wantErr int) func(t *testing.T) {
		return func(t *testing.T) {
			did := false
			fatal := false
			requested := false
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				_, _ = rw.Write([]byte("OK"))
			}))
			defer server.Close()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			switch wantErr {
			case 1:
				oldLogFatalf := logFatalf
				oldRequest := httpNewRequestWithContext
				defer func() {
					_ = recover()
					httpNewRequestWithContext = oldRequest
					logFatalf = oldLogFatalf
				}()
				httpNewRequestWithContext = func(context.Context, string, string, io.Reader) (*http.Request, error) {
					requested = true
					return nil, errors.New("request error")
				}
				logFatalf = func(msg string, v ...interface{}) {
					fatal = true
					panic(fmt.Sprintf(msg, v))
				}
			case 2:
				oldHttpClientDo := httpClientDo
				oldLogFatalf := logFatalf
				defer func() {
					_ = recover()
					httpClientDo = oldHttpClientDo
					logFatalf = oldLogFatalf
				}()
				httpClientDo = func(*http.Request) (*http.Response, error) {
					did = true
					return nil, errors.New("client error")
				}
				logFatalf = func(msg string, v ...interface{}) {
					fatal = true
					panic(fmt.Sprintf(msg, v))
				}
			}

			resp := MustPostWithContext(ctx, server.URL, "application/json", strings.NewReader("{}"))

			switch wantErr {
			case 0:
				defer resp.Body.Close()
				body, _ := ioutil.ReadAll(resp.Body)
				if string(body) != "OK" {
					t.Errorf("content mismatch: %s", body)
				}
			case 1:
				if !requested || did || !fatal {
					t.Errorf("wrong journey: %v, %v, %v", requested, did, fatal)
				}
			case 2:
				if !requested || !did || !fatal {
					t.Errorf("wrong journey: %v, %v, %v", requested, did, fatal)
				}
			}
		}
	}

	t.Run("with valid response", compare(0))
	t.Run("with request error", compare(1))
	t.Run("with client error", compare(2))
}
