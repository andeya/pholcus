package goutil

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TarGz compresses and archives tar.gz file.
func TarGz(src, dst string, includePrefix bool, logOutput func(string, ...interface{}), ignoreElem ...string) (err error) {
	// Create dst file
	fw, err := os.Create(dst)
	if err != nil {
		return
	}
	err = TarGzTo(src, fw, includePrefix, logOutput, ignoreElem...)
	fw.Close()
	if err != nil {
		os.Remove(dst)
	}
	return err
}

// TarGzTo compresses and archives tar.gz to dst writer.
func TarGzTo(src string, dstWriter io.Writer, includePrefix bool, logOutput func(string, ...interface{}), ignoreElem ...string) (err error) {
	src, err = filepath.Abs(src)
	if err != nil {
		return
	}
	srcFi, err := os.Stat(src)
	if err != nil {
		return
	}

	gw := gzip.NewWriter(dstWriter)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	var separator = string(filepath.Separator)

	var a = make([]string, 0, len(ignoreElem)+1)
	for _, v := range ignoreElem {
		v = strings.Trim(v, separator)
		if v == "" {
			continue
		}
		a = append(a, v)
	}
	ignoreElem = append(a, ".DS_Store")

	var prefix string
	if !srcFi.IsDir() || includePrefix {
		prefix, _ = filepath.Split(src)
	} else {
		prefix = src + separator
	}

	return filepath.Walk(src, func(fileName string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}

		// Because hdr.Name is base name,
		// once packaged, all files will pile up and destroy the original directory structure.
		hdr.Name = strings.TrimPrefix(fileName, prefix)

		// ignore files
		for _, v := range ignoreElem {
			if hdr.Name == v ||
				strings.HasPrefix(hdr.Name, v+separator) ||
				strings.HasSuffix(hdr.Name, separator+v) ||
				strings.Contains(hdr.Name, separator+v+separator) {
				return nil
			}
		}

		// If it is not a standard file, it will not be processed, such as a directory.
		if !fi.Mode().IsRegular() {
			return nil
		}

		// write file infomation
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		fr, err := os.Open(fileName)
		defer fr.Close()
		if err != nil {
			return err
		}

		n, err := io.Copy(tw, fr)
		if err != nil {
			return err
		}
		if logOutput != nil {
			logOutput("tar.gz: packaged %s, written %d bytes\n", hdr.Name, n)
		}
		return nil
	})
}
