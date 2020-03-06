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
	content []byte
	count   int
	err     int
}

func (w *writer) Bytes() []byte {
	return w.content
}

func (w *writer) Write(buf []byte) (int, error) {
	w.count++

	if w.err == 1 && w.count == 1 {
		return 0, errors.New("header write error")
	} else if w.err == 2 && w.count == 2 {
		return 0, errors.New("message write error")
	}

	w.content = append(w.content, buf...)
	return len(buf), nil
}

func TestHostInit(t *testing.T) {
	t.Parallel()

	exec, _ := os.Executable()
	absExec, _ := filepath.EvalSymlinks(exec)

	cases := []struct{
		name string
		got  *Host
		want *Host
	}{
		{
			"with default",
			(&Host{}).Init(),
			&Host{
				AppName: "native-messaging-host",
				AppDesc: "native-messaging-host",
				AppType: "stdio",
				AutoUpdate: false,
				ExecName: absExec,
				ByteOrder: binary.LittleEndian,
			},
		},
		{
			"with AppName",
			(&Host{
				AppName: "my.app.name",
			}).Init(),
			&Host{
				AppName: "my.app.name",
				AppDesc: "my.app.name",
				AppType: "stdio",
				AutoUpdate: false,
				ExecName: absExec,
				ByteOrder: binary.LittleEndian,
			},
		},
		{
			"with AppName, AppDesc",
			(&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
			}).Init(),
			&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "stdio",
				AutoUpdate: false,
				ExecName: absExec,
				ByteOrder: binary.LittleEndian,
			},
		},
		{
			"with AppName, AppDesc, AppType",
			(&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "any",
			}).Init(),
			&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "any",
				AutoUpdate: false,
				ExecName: absExec,
				ByteOrder: binary.LittleEndian,
			},
		},
		{
			"with AppName, AppDesc, AppType, ByteOrder",
			(&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "any",
				ByteOrder: binary.BigEndian,
			}).Init(),
			&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "any",
				AutoUpdate: false,
				ExecName: absExec,
				ByteOrder: binary.BigEndian,
			},
		},
		{
			"with AppName, AppDesc, AppType, ByteOrder, UpdateUrl",
			(&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "any",
				ByteOrder: binary.BigEndian,
				UpdateUrl: "https://www.google.com",
			}).Init(),
			&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "any",
				AutoUpdate: false,
				ExecName: absExec,
				ByteOrder: binary.BigEndian,
				UpdateUrl: "https://www.google.com",
			},
		},
		{
			"with AppName, AppDesc, AppType, ByteOrder, UpdateUrl, Version",
			(&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "any",
				ByteOrder: binary.BigEndian,
				UpdateUrl: "https://www.google.com",
				Version: "0.0.0",
			}).Init(),
			&Host{
				AppName: "my.app.name",
				AppDesc: "Description of my app",
				AppType: "any",
				AutoUpdate: true,
				ExecName: absExec,
				ByteOrder: binary.BigEndian,
				UpdateUrl: "https://www.google.com",
				Version: "0.0.0",
			},
		},
	}

	for _, tc := range cases {
		tc := tc // https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tc.want, tc.got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHostOnMessage(t *testing.T) {
	t.Parallel()

	cases := []struct{
		name   string
		err    bool
		exit   bool
		reader io.Reader
		want   *H
	}{
		{
			"with nothing",
			true,
			true,
			bytes.NewReader([]byte("")),
			&H{},
		},
		{
			"with empty message",
			false,
			false,
			getReader(t, ""),
			&H{},
		},
		{
			"with empty object",
			false,
			false,
			getReader(t, "{}"),
			&H{},
		},
		{
			"with valid object",
			false,
			false,
			getReader(t, `{"key":"value"}`),
			&H{"key": "value"},
		},
		{
			"with invalid object",
			true,
			false,
			getReader(t, `{"key":"value}`),
			&H{},
		},
	}

	for _, tc := range cases {
		tc := tc // https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			exited := false

			if tc.exit {
				oldRuntimeGoexit := runtimeGoexit
				defer func() { runtimeGoexit = oldRuntimeGoexit }()
				runtimeGoexit = func() { exited = true }
			}

			var buf bytes.Buffer
			got := &H{}
			tee := io.TeeReader(tc.reader, &buf)

			if err := (&Host{
				ByteOrder: binary.LittleEndian,
			}).OnMessage(tee, got); !tc.err && err != nil {
				input, _ := ioutil.ReadAll(&buf)
				t.Fatalf("got error %s: %s", input, err)
			} else if tc.err && err == nil {
				input, _ := ioutil.ReadAll(&buf)
				t.Fatalf("want error: %s", input)
			} else if tc.exit && !exited {
				input, _ := ioutil.ReadAll(&buf)
				t.Fatalf("want exit: %s", input)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHostPostMessage(t *testing.T) {
	t.Parallel()

	cases := []struct{
		name    string
		err     bool
		message interface{}
		want    *H
		writer  *writer
	}{
		{
			"with invalid object",
			true,
			make(chan int),
			&H{},
			&writer{},
		},
		{
			"with header writer error",
			true,
			&H{},
			&H{},
			&writer{err: 1},
		},
		{
			"with message writer error",
			true,
			&H{},
			&H{},
			&writer{err: 2},
		},
		{
			"with empty object",
			false,
			&H{},
			&H{},
			&writer{},
		},
		{
			"with valid object",
			false,
			&H{"key": "value"},
			&H{"key": "value"},
			&writer{},
		},
	}

	for _, tc := range cases {
		tc := tc // https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := &H{}

			if err := (&Host{
				ByteOrder: binary.LittleEndian,
			}).PostMessage(tc.writer, tc.message); !tc.err && err != nil {
				t.Fatalf("got error %+v: %s", tc.message, err)
			} else if tc.err && err == nil {
				t.Fatalf("want error: %v", tc.message)
			}

			if !tc.err {
				if err := json.Unmarshal(tc.writer.Bytes()[4:], got); err != nil {
					t.Fatalf("unmarshal error %v: %s", tc.message, err)
				}
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func getReader(t *testing.T, message string) io.Reader {
	t.Helper()
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, (uint32)(len(message)))
	return bytes.NewReader(append(buf, []byte(message)...))
}
