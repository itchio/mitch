package mitch

import (
	"archive/zip"
	"bytes"
	"io"
)

type ArchiveContext struct {
	Name    string
	Entries map[string]*ArchiveEntry
}

func (ac *ArchiveContext) SetName(name string) {
	ac.Name = name
}

func (ac *ArchiveContext) Entry(path string) *ArchiveEntry {
	entry := &ArchiveEntry{path: path}
	ac.Entries[path] = entry
	return entry
}

func (ac *ArchiveContext) CompressZip() []byte {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for path, e := range ac.Entries {
		w, err := zw.Create(path)
		must(err)
		_, err = io.Copy(w, bytes.NewReader(e.buf.Bytes()))
		must(err)
	}
	must(zw.Close())

	return buf.Bytes()
}

type ArchiveEntry struct {
	path string
	buf  bytes.Buffer
}

func (ae *ArchiveEntry) String(s string) {
	ae.buf.WriteString(s)
}

func (ae *ArchiveEntry) Bytes(p []byte) {
	ae.buf.Write(p)
}
