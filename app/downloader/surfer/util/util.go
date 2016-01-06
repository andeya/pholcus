// Package util contains some utility methods used by other packages.
package util

import (
	"fmt"
	"hash/crc32"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// 返回编码后的url.URL指针、及解析错误
func UrlEncode(urlStr string) (*url.URL, error) {
	urlObj, err := url.Parse(urlStr)
	urlObj.RawQuery = urlObj.Query().Encode()
	return urlObj, err
}

// 制作特征值
func MakeHash(s string) string {
	const IEEE = 0xedb88320
	var IEEETable = crc32.MakeTable(IEEE)
	hash := fmt.Sprintf("%x", crc32.Checksum([]byte(s), IEEETable))
	return hash
}

// 创建目录
func Mkdir(Path string) error {
	p, _ := path.Split(Path)
	if p == "" {
		return nil
	}
	d, err := os.Stat(p)
	if err != nil || !d.IsDir() {
		if err = os.MkdirAll(p, 0777); err != nil {
			log.Printf("创建路径失败[%v]: %v\n", Path, err)
			return err
		}
	}
	return nil
}

// The GetWDPath gets the work directory path.
func GetWDPath() string {
	wd := os.Getenv("GOPATH")
	if wd == "" {
		panic("GOPATH is not setted in env.")
	}
	return wd
}

// The IsDirExists judges path is directory or not.
func IsDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}

	panic("util isDirExists not reached")
}

// The IsFileExists judges path is file or not.
func IsFileExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return !fi.IsDir()
	}

	panic("util isFileExists not reached")
}

// 遍历并返回指定类型范围的文件名列表
// 默认返回所有文件
func WalkFiles(path string, suffixes ...string) (filelist []string) {
	path, _ = filepath.Abs(path)

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if len(suffixes) == 0 {
			filelist = append(filelist, path)
			return nil
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(path, suffix) {
				filelist = append(filelist, path)
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("filepath.Walk() returned %v\n", err)
	}

	return
}
