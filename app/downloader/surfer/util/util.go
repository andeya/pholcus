// Package util contains some utility methods used by other packages.
package util

import (
	"fmt"
	"hash/crc32"
	"net/url"
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
