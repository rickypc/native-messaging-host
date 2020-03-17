// zip.go - Zip related functionality.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package packer

import (
	"archive/zip"
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Unzip reads the zip-compressed file from reader and writes it into target dir.
func Unzip(r io.Reader, dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("unzip mkdir -p %s error: %v", dir, err)
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		log.Fatalf("download zip error: %v", err)
	}

	b := bytes.NewReader(buf.Bytes())
	zr, err := zip.NewReader(b, int64(b.Len()))
	if err != nil {
		log.Fatalf("open zip error: %v", err)
	}

	for _, f := range zr.File {
		name := filepath.Join(dir, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(name, f.Mode()); err != nil {
				log.Fatalf("unzip mkdir -p %s error: %v", name, err)
			}
			continue
		}

		unzipEntry(f, name)
	}
}

// unzipEntry creates new file or folder on given zip file entry.
func unzipEntry(f *zip.File, name string) {
	src, err := f.Open()
	if err != nil {
		log.Fatalf("unzip open file error: %v", err)
	}
	defer src.Close()

	dst, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode())
	if err != nil {
		log.Fatalf("unzip create file error: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Fatalf("unzip write file error: %v", err)
	}
}
