// tar.go - Tar related functionality.
// Copyright (c) 2018 - 2020  Richard Huang <rickypc@users.noreply.github.com>
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package packer

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// removeLink is a wrapper to remove given path and log any error.
func removeLink(name string) {
	if _, err := os.Lstat(name); err == nil {
		if err := osRemove(name); err != nil {
			logFatalf("untar rm %s error: %v", name, err)
		}
	}
}

// Untar reads the gzip-compressed tar file from reader and writes it into
// target dir.
func Untar(r io.Reader, dir string) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		log.Fatalf("gunzip error: %v", err)
	}
	defer zr.Close()

	tr := tar.NewReader(zr)
	for {
		if h, err := tr.Next(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalf("untar error: %v", err)
			}
		} else if h != nil {
			if !validRelPath(h.Name) {
				log.Fatalf("untar invalid name: %q", h.Name)
			}
			untarEntry(tr, h, dir)
		}
	}
}

// untarEntry creates new file or folder on given tar header.
func untarEntry(tr *tar.Reader, h *tar.Header, dir string) {
	mode := h.FileInfo().Mode()
	name := filepath.Join(dir, filepath.FromSlash(h.Name))

	switch h.Typeflag {
	case tar.TypeDir:
		if err := os.MkdirAll(name, mode); err != nil {
			log.Fatalf("untar mkdir -p %s error: %v", name, err)
		}
	case tar.TypeReg, tar.TypeRegA:
		file, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
		if err != nil {
			log.Fatalf("untar create %s error: %v", name, err)
		}

		n, err := io.Copy(file, tr)
		if closeErr := file.Close(); closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				err = fmt.Errorf("%w %v", err, closeErr)
			}
		}

		if err != nil {
			log.Fatalf("untar write %s error: %v", name, err)
		}

		if n != h.Size {
			log.Fatalf("wrote %s only %d bytes of %d", name, n, h.Size)
		}
	case tar.TypeLink:
		removeLink(name)
		if err := os.Link(filepath.Join(dir, h.Linkname), name); err != nil {
			log.Fatalf("untar ln %s: %v", name, err)
		}
	case tar.TypeSymlink:
		removeLink(name)
		if err := os.Symlink(h.Linkname, name); err != nil {
			log.Fatalf("untar ln -s %s: %v", name, err)
		}
	case tar.TypeBlock, tar.TypeChar, tar.TypeFifo, tar.TypeGNUSparse, tar.TypeXGlobalHeader:
		break
	default:
		log.Fatalf("untar unknown type %s: %s", mode, name)
	}
}

// validRelPath validates given relative path.
func validRelPath(p string) bool {
	if p == "" || strings.Contains(p, `\`) || strings.HasPrefix(p, "/") || strings.Contains(p, "../") {
		return false
	}
	return true
}
