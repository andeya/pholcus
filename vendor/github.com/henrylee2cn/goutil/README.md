# goutil [![report card](https://goreportcard.com/badge/github.com/henrylee2cn/goutil?style=flat-square)](http://goreportcard.com/report/henrylee2cn/goutil) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/goutil)

Common and useful utils for the Go project development.

## 1. Inclusion criteria

- Only rely on the Go standard package
- Functions or lightweight packages
- Non-business related general tools

## 2. Contents

- [Tools](#) Some useful small functions.
- [BitSet](bitset) A bit set
- [Calendar](calendar) Chinese Lunar Calendar, Solar Calendar and cron time rules
- [Cmder](cmder) Cmder exec cmd and catch the result
- [CoarseTime](coarsetime) Current time truncated to the nearest 100ms
- [Errors](errors) Improved errors package.
- [Graceful](graceful) Shutdown or reboot current process gracefully
- [HTTPBody](httpbody) HTTP body builder
- [Password](password) Check password
- [GoPool](pool) Goroutines' pool
- [ResPool](pool) Resources' pool
- [Workshop](pool) Non-blocking asynchronous multiplex resource pool
- [Status](status) A handling status with code, msg, cause and stack
- [Tpack](tpack) Go underlying type data
- [Versioning](versioning) Version comparison tool that conforms to semantic version 2.0.0
