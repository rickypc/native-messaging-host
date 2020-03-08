// host.go - Native messaging host config, message handler, and else.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package host provides native-messaging host configurations, send and receive
// message handler, manifest install and uninstall, as well as auto update daily
// check.
//
// * Sending Message
//
//   messaging := (&host.Host{}).Init()
//
//   // host.H is a shortcut to map[string]interface{}
//   response := &host.H{"key":"value"}
//
//   // Write message from response to os.Stdout.
//   if err := messaging.PostMessage(os.Stdout, response); err != nil {
//     log.Fatalf("messaging.PostMessage error: %v", err)
//   }
//
//   // Log response.
//   log.Printf("response: %+v", response)
//
// * Receiving Message
//
//   // Ensure func main returned after calling runtime.Goexit
//   // See https://golang.org/pkg/runtime/#Goexit.
//   defer os.Exit(0)
//
//   messaging := (&host.Host{}).Init()
//
//   // host.H is a shortcut to map[string]interface{}
//   request := &host.H{}
//
//   // Read message from os.Stdin to request.
//   if err := messaging.OnMessage(os.Stdin, request); err != nil {
//     log.Fatalf("messaging.OnMessage error: %v", err)
//   }
//
//   // Log request.
//   log.Printf("request: %+v", request)
//
// * Install and Uninstall Hooks
//
//   // AllowedExts is a list of extensions that should have access to the native messaging host.
//   // See [native messaging manifest][7]
//   messaging := (&host.Host{
//     AppName:     "tld.domain.sub.app.name",
//     AllowedExts: []string{"chrome-extension://XXX/", "chrome-extension://YYY/"},
//   }).Init()
//
//   ...
//
//   // When you need to install.
//   if err := messaging.Install(); err != nil {
//     log.Printf("install error: %v", err)
//   }
//
//   ...
//
//   // When you need to uninstall.
//   host.Uninstall()
//
// * Auto Update Configuration
//
//   // updates.xml example for cross platform executable:
//   <?xml version='1.0' encoding='UTF-8'?>
//   <gupdate xmlns='http://www.google.com/update2/response' protocol='2.0'>
//     <app appid='tld.domain.sub.app.name'>
//       <updatecheck codebase='https://sub.domain.tld/app.download.all' version='1.0.0' />
//     </app>
//   </gupdate>
//
//   // updates.xml example for individual platform executable:
//   <?xml version='1.0' encoding='UTF-8'?>
//   <gupdate xmlns='http://www.google.com/update2/response' protocol='2.0'>
//     <app appid='tld.domain.sub.app.name'>
//       <updatecheck codebase='https://sub.domain.tld/app.download.darwin' os='darwin' version='1.0.0' />
//       <updatecheck codebase='https://sub.domain.tld/app.download.linux' os='linux' version='1.0.0' />
//       <updatecheck codebase='https://sub.domain.tld/app.download.exe' os='windows' version='1.0.0' />
//     </app>
//   </gupdate>
//
//   // It will do daily update check.
//   messaging := (&host.Host{
//     AppName:   "tld.domain.sub.app.name",
//     UpdateUrl: "https://sub.domain.tld/updates.xml", // It follows [update manifest][2]
//     Version:   "1.0.0",                              // Current version, it must follow [SemVer][6]
//   }).Init()
package host

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

// ioutilWriteFile is a shortcut to ioutil.WriteFile. It helps write testable code.
var ioutilWriteFile = ioutil.WriteFile

// osMkdirAll is a shortcut to os.MkdirAll. It helps write testable code.
var osMkdirAll = os.MkdirAll

// runtimeGoexit is a shortcut to runtime.Goexit. It helps write testable code.
var runtimeGoexit = runtime.Goexit

// H is a map[string]interface{} type shortcut and represents a dynamic
// key-value-pair data.
type H map[string]interface{}

// Host represents a single native messaging host, where all native messaging
// host operations can be done.
type Host struct {
	AppName     string           `json:"name"`
	AppDesc     string           `json:"description"`
	ExecName    string           `json:"path"`
	AppType     string           `json:"type"`
	AllowedExts []string         `json:"allowed_origins"`
	AutoUpdate  bool             `json:"-"`
	ByteOrder   binary.ByteOrder `json:"-"`
	UpdateUrl   string           `json:"-"`
	Version     string           `json:"-"`
}

// Init sets default value to its fields and return the Host pointer back.
//
// * AppName is an application name in manifest file and will be defaulted to
// current executable file name without extension, if any.
//
// * AppDesc is an application description in manifest file and will be defaulted
// to current AppName.
//
// * AppType is an application communication type in manifest file and will be
// defaulted to "stdio".
//
// * AutoUpdate indicates whether update check will be perform for this
// application and will be defaulted to true only if UpdateUrl and application
// Version are present, otherwise it will be false.
//
// * ByteOrder specifies how to convert byte sequences into unsigned integers and
// will be defaulted to binary.LittleEndian.
//
// * ExecName is an executable path used across the module and will get assigned
// to current executable's absolute path after the evaluation of any symbolic
// links.
//
//   messaging := (&host.Host{}).Init()
func (h *Host) Init() *Host {
	exec, _ := os.Executable()
	evaled, _ := filepath.EvalSymlinks(exec)
	h.ExecName, _ = filepath.Abs(evaled)

	if h.AppName == "" {
		h.AppName = strings.TrimSuffix(filepath.Base(h.ExecName), path.Ext(h.ExecName))
	}

	if h.AppDesc == "" {
		h.AppDesc = h.AppName
	}

	if h.AppType == "" {
		h.AppType = "stdio"
	}

	if h.ByteOrder == nil {
		h.ByteOrder = binary.LittleEndian
	}

	if h.UpdateUrl != "" && h.Version != "" {
		h.AutoUpdate = true
	}

	return h
}

// OnMessage reads message header and message body from given reader and
// unmarshal to given struct. It will return error when it come across one.
//
//   // Ensure func main returned after calling runtime.Goexit
//   // See https://golang.org/pkg/runtime/#Goexit.
//   defer os.Exit(0)
//
//   messaging := (&host.Host{}).Init()
//
//   // host.H is a shortcut to map[string]interface{}
//   request := &host.H{}
//
//   // Read message from os.Stdin to request.
//   if err := messaging.OnMessage(os.Stdin, request); err != nil {
//     log.Fatalf("messaging.OnMessage error: %v", err)
//   }
//
//   // Log request.
//   log.Printf("request: %+v", request)
func (h *Host) OnMessage(reader io.Reader, v interface{}) error {
	length, err := h.readHeader(reader)

	if err != nil {
		return err
	}

	// Nothing to read.
	if length == 0 {
		return nil
	}

	// Read message body.
	if err := json.NewDecoder(io.LimitReader(reader, int64(length))).Decode(v); err != nil {
		return err
	}

	return nil
}

// readHeader reads message header and will return the message length. It will
// return error when it come across one.
func (h *Host) readHeader(reader io.Reader) (uint32, error) {
	// Read message length.
	var length uint32

	if err := binary.Read(reader, h.ByteOrder, &length); err != nil {
		if err == io.EOF {
			h.AutoUpdateCheck()

			// Exit gracefully.
			runtimeGoexit()
		}

		return length, err
	}

	return length, nil
}

// PostMessage marshals given struct and writes message header and message body
// to given writer. It will return error when it come across one.
//
//   messaging := (&host.Host{}).Init()
//
//   // host.H is a shortcut to map[string]interface{}
//   response := &host.H{"key":"value"}
//
//   // Write message from response to os.Stdout.
//   if err := messaging.PostMessage(os.Stdout, response); err != nil {
//     log.Fatalf("messaging.PostMessage error: %v", err)
//   }
//
//   // Log response.
//   log.Printf("response: %+v", response)
func (h *Host) PostMessage(writer io.Writer, v interface{}) error {
	message, err := json.Marshal(v)
	if err != nil {
		return err
	}

	length := len(message)

	if err := h.writeHeader(writer, length); err != nil {
		return err
	}

	// Write message body.
	if n, err := writer.Write(message); err != nil || n != length {
		return err
	}

	return nil
}

// writeHeader writes message length into message header. It will return error
// when it come across one.
func (h *Host) writeHeader(writer io.Writer, length int) error {
	header := make([]byte, 4)
	h.ByteOrder.PutUint32(header, (uint32)(length))

	if n, err := writer.Write(header); err != nil || n != len(header) {
		return err
	}

	return nil
}
