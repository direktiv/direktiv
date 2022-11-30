package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func safeLoadFile(filePath string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	if filePath == "" {
		// skip if filePath is empty
		return buf, nil
	}

	fStat, err := os.Stat(filePath)
	if err != nil {
		return buf, err
	}

	if fStat.Size() > maxSize {
		return buf, fmt.Errorf("file is larger than maximum allowed size: %v. Set configfile 'max-size' to change", maxSize)
	}

	fData, err := os.ReadFile(filePath)
	if err != nil {
		return buf, err
	}

	buf = bytes.NewBuffer(fData)

	return buf, nil
}

func safeLoadStdIn() (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)

	fi, err := os.Stdin.Stat()
	if err != nil {
		return buf, err
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		// No stdin
		return buf, nil
	}

	if fi.Size() > maxSize {
		return buf, fmt.Errorf("stdin is larger than maximum allowed size: %v. Set configfile 'max-size' to change", maxSize)
	}

	fData, err := io.ReadAll(os.Stdin)
	if err != nil {
		return buf, err
	}

	buf = bytes.NewBuffer(fData)

	return buf, nil
}
