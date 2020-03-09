// download_test.go - Test for download related functionality.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package host

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var opened bool

type StubFileSystem struct {
	io.Writer
}

func (s *StubFileSystem) Close() error {
	return nil
}

func (s *StubFileSystem) OpenFile(name string, flag int, perm os.FileMode) (FileInterface, error) {
	opened = true
	return s, nil
}

type StubErrorFileSystem struct {
	io.Writer
}

func (s *StubErrorFileSystem) Close() error {
	return nil
}

func (s *StubErrorFileSystem) OpenFile(name string, flag int, perm os.FileMode) (FileInterface, error) {
	opened = true
	return nil, errors.New("open file error")
}

func TestDownloadLatest(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	compare := func(wantErr int, want *H) func(t *testing.T) {
		return func(t *testing.T) {
			copied := false
			opened = false
			renamed := 0
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				if wantErr == 2 {
					rw.WriteHeader(http.StatusNotFound)
					_, _ = rw.Write([]byte(http.StatusText(http.StatusNotFound)))
				} else {
					_, _ = rw.Write([]byte("OK"))
				}
			}))
			defer server.Close()
			targetName := "testdata/down"
			url := server.URL

			switch wantErr {
			case 0:
				if err := ioutil.WriteFile(targetName, []byte(""), 0644); err != nil {
					t.Fatalf("touch file error: %v", err)
				}
				defer func() { os.Remove(targetName) }()
			case 1:
				oldFs := fs
				oldIoCopy := ioCopy
				oldOsRename := osRename
				defer func() {
					fs = oldFs
					ioCopy = oldIoCopy
					osRename = oldOsRename
				}()
				fs = &StubFileSystem{bytes.NewBufferString("")}
				ioCopy = func(io.Writer, io.Reader) (int64, error) {
					copied = true
					return 0, nil
				}
				osRename = func(string, string) error { renamed++; return nil }
			case 3:
				oldOsRename := osRename
				defer func() {
					osRename = oldOsRename
				}()
				osRename = func(string, string) error {
					renamed++
					return errors.New("backup error")
				}
			case 4:
				oldFs := fs
				oldOsRename := osRename
				defer func() {
					fs = oldFs
					osRename = oldOsRename
				}()
				fs = &StubErrorFileSystem{bytes.NewBufferString("")}
				osRename = func(string, string) error { renamed++; return nil }
			case 5:
				oldFs := fs
				oldOsRename := osRename
				defer func() {
					fs = oldFs
					osRename = oldOsRename
				}()
				fs = &StubErrorFileSystem{bytes.NewBufferString("")}
				osRename = func(string, string) error {
					renamed++
					if renamed == 2 {
						return errors.New("open file revert error")
					} else {
						return nil
					}
				}
			case 6:
				oldFs := fs
				oldIoCopy := ioCopy
				oldOsRename := osRename
				defer func() {
					fs = oldFs
					ioCopy = oldIoCopy
					osRename = oldOsRename
				}()
				fs = &StubFileSystem{bytes.NewBufferString("")}
				ioCopy = func(io.Writer, io.Reader) (int64, error) {
					copied = true
					return 0, errors.New("download error")
				}
				osRename = func(string, string) error {
					renamed++
					return nil
				}
			case 7:
				oldFs := fs
				oldIoCopy := ioCopy
				oldOsRename := osRename
				defer func() {
					fs = oldFs
					ioCopy = oldIoCopy
					osRename = oldOsRename
				}()
				fs = &StubFileSystem{bytes.NewBufferString("")}
				ioCopy = func(io.Writer, io.Reader) (int64, error) {
					copied = true
					return 0, errors.New("download error")
				}
				osRename = func(string, string) error {
					renamed++
					if renamed == 2 {
						return errors.New("open file revert error")
					} else {
						return nil
					}
				}
			}

			if err := (&Host{ExecName: targetName}).downloadLatest(url); wantErr < 2 && err != nil {
				t.Errorf("download error: %v", err)
			} else if wantErr > 1 && err == nil {
				t.Fatal("want error")
			}

			if wantErr == 0 {
				if info, err := os.Stat(targetName); err != nil {
					t.Fatalf("missing file: %v", err)
				} else if fmt.Sprintf("%#o", info.Mode().Perm()) != "0755" {
					t.Fatalf("wrong file permission: %v", err)
				}

				if buf, err := ioutil.ReadFile(targetName); err != nil {
					t.Fatalf("file read error: %v", err)
				} else if string(buf) != "OK" {
					t.Fatal("wrong content")
				}
			}

			got := &H{"copied": copied, "opened": opened, "renamed": renamed}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch for %d (-want +got):\n%s", wantErr, diff)
			}
		}
	}

	t.Run("with download latest on fs", compare(0, &H{"copied": false, "opened": false, "renamed": 0}))
	t.Run("with download latest", compare(1, &H{"copied": true, "opened": true, "renamed": 1}))
	t.Run("with non-OK status code error", compare(2, &H{"copied": false, "opened": false,
		"renamed": 0}))
	t.Run("with create backup error", compare(3, &H{"copied": false, "opened": false,
		"renamed": 1}))
	t.Run("with create file error", compare(4, &H{"copied": false, "opened": true,
		"renamed": 2}))
	t.Run("with create file revert error", compare(5, &H{"copied": false, "opened": true,
		"renamed": 2}))
	t.Run("with download file error", compare(6, &H{"copied": true, "opened": true,
		"renamed": 2}))
	t.Run("with download revert error", compare(7, &H{"copied": true, "opened": true,
		"renamed": 2}))
}

func TestDownloadUrlAndVersion(t *testing.T) {
	t.Parallel()

	log.SetOutput(ioutil.Discard)

	compare := func(wantErr int, want *H) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				xml := `<?xml version='1.0' encoding='UTF-8'?>
<gupdate xmlns='http://www.google.com/update2/response' protocol='2.0'>
  <app appid='tld.domain.sub.app.name'>
    <updatecheck codebase='https://sub.domain.tld/app.download.all' version='1.0.0' />
  </app>
</gupdate`

				if wantErr != 1 {
					xml += ">"
				}

				_, _ = rw.Write([]byte(xml))
			}))
			defer server.Close()

			h := &Host{UpdateUrl: server.URL}
			if wantErr != 2 {
				h.AppName = "tld.domain.sub.app.name"
			}

			url, version, err := h.getDownloadUrlAndVersion()
			got := &H{"err": err, "url": url, "version": version}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		}
	}

	t.Run("with valid response", compare(0, &H{"err": nil,
		"url": "https://sub.domain.tld/app.download.all", "version": "1.0.0"}))
	t.Run("with xml decoder error", compare(1, &H{
		"err": &xml.SyntaxError{Line: 6, Msg: "unexpected EOF"}, "url": "", "version": ""}))
	t.Run("with AppName mismatch", compare(2, &H{"err": nil, "url": "", "version": ""}))
}
