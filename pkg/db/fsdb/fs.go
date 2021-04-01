package fsdb

/*
Filesystem based database
*/

import (
	"errors"
	"fmt"
	"github.com/raspi/torjuja/pkg/db/iface"
	"os"
	"path"
	"strings"
)

// Check implementation
var _ iface.Database = FileSystemDB{}

type FileSystemDB struct {
	basepath          string
	allowedPath       string
	defaultPermission os.FileMode
}

func New(basepath string) (*FileSystemDB, error) {
	if !path.IsAbs(basepath) {
		return nil, fmt.Errorf(`not absolute path: %q`, basepath)
	}

	return &FileSystemDB{
		basepath:          basepath,
		allowedPath:       path.Join(basepath, `allowed`),
		defaultPermission: 0660,
	}, nil
}

func (f FileSystemDB) getPath(name string, t string) string {
	return path.Join(f.allowedPath, t, strings.Join(reverse(strings.Split(name, `.`)), string(os.PathSeparator)))
}

func (f FileSystemDB) allowed(name string, t string) (bool, error) {
	switch t {
	case `A`, `AAAA`:
		t = `IP`
	}

	fi, err := os.Stat(path.Join(f.getPath(name, t), `allow`))

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return false, err
	}

	return fi.Mode().IsRegular(), nil
}

func (f FileSystemDB) AllowedA(name string) (bool, error) {
	return f.allowed(name, `A`)
}

func (f FileSystemDB) AllowedAAAA(name string) (bool, error) {
	return f.allowed(name, `AAAA`)
}

func (f FileSystemDB) AllowedPTR(name string) (bool, error) {
	return f.allowed(name, `PTR`)
}

func (f FileSystemDB) allow(name string, t string) error {
	switch t {
	case `A`, `AAAA`:
		t = `IP`
	}

	fpath := f.getPath(name, t)

	err := os.MkdirAll(fpath, f.defaultPermission)
	if err != nil {
		return err
	}

	fh, err := os.Create(path.Join(fpath, `allow`))
	if err != nil {
		return err
	}
	defer fh.Close()

	err = fh.Chmod(f.defaultPermission)
	if err != nil {
		return err
	}

	return nil
}

func (f FileSystemDB) AllowA(name string) error {
	return f.allow(name, `A`)
}

func (f FileSystemDB) AllowAAAA(name string) error {
	return f.allow(name, `AAAA`)
}

func (f FileSystemDB) AllowPTR(name string) error {
	return f.allow(name, `PTR`)
}

func reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}
