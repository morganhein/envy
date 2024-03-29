package test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	//make folder, if it doesn't exist
	dstFolder := filepath.Dir(dst)
	err = os.MkdirAll(dstFolder, os.ModeDir)
	if err != nil {
		return 0, err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func exists(target string) (bool, error) {
	if _, err := os.Stat(target); err == nil {
		// path/to/whatever exists
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		// path/to/whatever does *not* exist
		return false, nil
	} else {
		return false, err
	}
}

func newCtx(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
