// host_test.go - Native messaging host tests.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package host

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/google/go-cmp/cmp"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type writer struct {
	content    []byte
	count, err int
}

func (w *writer) Bytes() []byte {
	return w.content
}

func (w *writer) Write(buf []byte) (int, error) {
	w.count++

	if w.count == w.err {
		if w.err == 1 {
			return 0, errors.New("header write error")
		} else if w.err == 2 {
			return 0, errors.New("message write error")
		}
	}

	w.content = append(w.content, buf...)
	return len(buf), nil
}

func TestHostInit(t *testing.T) {
	t.Parallel()

	exec, _ := os.Executable()
	absExec, _ := filepath.EvalSymlinks(exec)

	compare := func(got *Host, want *Host) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		}
	}

	t.Run("with default", compare((&Host{}).Init(), &Host{
		AppName:    "native-messaging-host",
		AppDesc:    "native-messaging-host",
		AppType:    "stdio",
		AutoUpdate: false,
		ExecName:   absExec,
		ByteOrder:  binary.LittleEndian,
	}))

	t.Run("with AppName", compare((&Host{
		AppName: "my.app.name",
	}).Init(), &Host{
		AppName:    "my.app.name",
		AppDesc:    "my.app.name",
		AppType:    "stdio",
		AutoUpdate: false,
		ExecName:   absExec,
		ByteOrder:  binary.LittleEndian,
	}))

	t.Run("with AppName, AppDesc", compare((&Host{
		AppName: "my.app.name",
		AppDesc: "Description of my app",
	}).Init(), &Host{
		AppName:    "my.app.name",
		AppDesc:    "Description of my app",
		AppType:    "stdio",
		AutoUpdate: false,
		ExecName:   absExec,
		ByteOrder:  binary.LittleEndian,
	}))

	t.Run("with AppName, AppDesc, AppType", compare((&Host{
		AppName: "my.app.name",
		AppDesc: "Description of my app",
		AppType: "any",
	}).Init(), &Host{
		AppName:    "my.app.name",
		AppDesc:    "Description of my app",
		AppType:    "any",
		AutoUpdate: false,
		ExecName:   absExec,
		ByteOrder:  binary.LittleEndian,
	}))

	t.Run("with AppName, AppDesc, AppType, ByteOrder", compare((&Host{
		AppName:   "my.app.name",
		AppDesc:   "Description of my app",
		AppType:   "any",
		ByteOrder: binary.BigEndian,
	}).Init(), &Host{
		AppName:    "my.app.name",
		AppDesc:    "Description of my app",
		AppType:    "any",
		AutoUpdate: false,
		ExecName:   absExec,
		ByteOrder:  binary.BigEndian,
	}))

	t.Run("with AppName, AppDesc, AppType, ByteOrder, UpdateUrl", compare((&Host{
		AppName:   "my.app.name",
		AppDesc:   "Description of my app",
		AppType:   "any",
		ByteOrder: binary.BigEndian,
		UpdateUrl: "https://www.google.com",
	}).Init(), &Host{
		AppName:    "my.app.name",
		AppDesc:    "Description of my app",
		AppType:    "any",
		AutoUpdate: false,
		ExecName:   absExec,
		ByteOrder:  binary.BigEndian,
		UpdateUrl:  "https://www.google.com",
	}))

	t.Run("with AppName, AppDesc, AppType, ByteOrder, UpdateUrl, Version", compare((&Host{
		AppName:   "my.app.name",
		AppDesc:   "Description of my app",
		AppType:   "any",
		ByteOrder: binary.BigEndian,
		UpdateUrl: "https://www.google.com",
		Version:   "0.0.0",
	}).Init(), &Host{
		AppName:    "my.app.name",
		AppDesc:    "Description of my app",
		AppType:    "any",
		AutoUpdate: true,
		ExecName:   absExec,
		ByteOrder:  binary.BigEndian,
		UpdateUrl:  "https://www.google.com",
		Version:    "0.0.0",
	}))
}

func TestHostOnMessage(t *testing.T) {
	t.Parallel()

	compare := func(wantErr bool, wantExit bool, message interface{}, want *H) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			exited := false
			got := &H{}
			reader := bytes.NewReader([]byte(""))

			if _, ok := message.(string); ok {
				header := make([]byte, 4)
				messageStr := message.(string)
				binary.LittleEndian.PutUint32(header, (uint32)(len(messageStr)))
				reader = bytes.NewReader(append(header, []byte(messageStr)...))
			}

			tee := io.TeeReader(reader, &buf)

			if wantExit {
				oldRuntimeGoexit := runtimeGoexit
				defer func() { runtimeGoexit = oldRuntimeGoexit }()
				runtimeGoexit = func() { exited = true }
			}

			if err := (&Host{
				ByteOrder: binary.LittleEndian,
			}).OnMessage(tee, got); !wantErr && err != nil {
				input, _ := ioutil.ReadAll(&buf)
				t.Fatalf("got error %s: %v", input, err)
			} else if wantErr && err == nil {
				input, _ := ioutil.ReadAll(&buf)
				t.Fatalf("want error: %s", input)
			} else if wantExit && !exited {
				input, _ := ioutil.ReadAll(&buf)
				t.Fatalf("want exit: %s", input)
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		}
	}

	t.Run("with nothing", compare(true, true, nil, &H{}))
	t.Run("with empty message", compare(false, false, "", &H{}))
	t.Run("with empty object", compare(false, false, "{}", &H{}))
	t.Run("with invalid object", compare(true, false, `{"key":"value}`, &H{}))
	t.Run("with valid object", compare(false, false, `{"key":"value"}`, &H{"key": "value"}))
}

func TestHostPostMessage(t *testing.T) {
	t.Parallel()

	compare := func(wantErr bool, message interface{}, want *H, writer *writer) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()

			got := &H{}

			if err := (&Host{
				ByteOrder: binary.LittleEndian,
			}).PostMessage(writer, message); !wantErr && err != nil {
				t.Fatalf("got error %+v: %v", message, err)
			} else if wantErr && err == nil {
				t.Fatalf("want error: %v", message)
			}

			if !wantErr {
				if err := json.Unmarshal(writer.Bytes()[4:], got); err != nil {
					t.Fatalf("unmarshal error %v: %v", message, err)
				}
			}

			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		}
	}

	t.Run("with invalid object", compare(true, make(chan int), &H{}, &writer{}))
	t.Run("with header writer error", compare(true, &H{}, &H{}, &writer{err: 1}))
	t.Run("with message writer error", compare(true, &H{}, &H{}, &writer{err: 2}))
	t.Run("with empty object", compare(false, &H{}, &H{}, &writer{}))
	t.Run("with valid object", compare(false, &H{"key": "value"}, &H{"key": "value"}, &writer{}))
}
