// Package data 提供了数据单元与文件单元的存储结构定义。
package data

import (
	"sync"
)

const (
	FieldRuleName     = "RuleName"
	FieldURL          = "Url"
	FieldParentURL    = "ParentUrl"
	FieldDownloadTime = "DownloadTime"
)

type (
	// DataCell is a storage unit for text data.
	DataCell map[string]interface{}
	// FileCell is a storage unit for file data.
	// Stored path format: file/"Dir"/"RuleName"/"time"/"Name"
	FileCell map[string]interface{}
)

var (
	dataCellPool = &sync.Pool{
		New: func() interface{} {
			return DataCell{}
		},
	}
	fileCellPool = &sync.Pool{
		New: func() interface{} {
			return FileCell{}
		},
	}
)

// GetDataCell returns a DataCell from the pool with the given fields.
func GetDataCell(ruleName string, data map[string]interface{}, url string, parentURL string, downloadTime string) DataCell {
	cell := dataCellPool.Get().(DataCell)
	cell[FieldRuleName] = ruleName
	cell["Data"] = data
	cell[FieldURL] = url
	cell[FieldParentURL] = parentURL
	cell[FieldDownloadTime] = downloadTime
	return cell
}

// GetFileCell returns a FileCell from the pool with the given fields.
func GetFileCell(ruleName, name string, bytes []byte) FileCell {
	cell := fileCellPool.Get().(FileCell)
	cell[FieldRuleName] = ruleName
	cell["Name"] = name
	cell["Bytes"] = bytes
	return cell
}

// PutDataCell returns a DataCell to the pool.
func PutDataCell(cell DataCell) {
	cell[FieldRuleName] = nil
	cell["Data"] = nil
	cell[FieldURL] = nil
	cell[FieldParentURL] = nil
	cell[FieldDownloadTime] = nil
	dataCellPool.Put(cell)
}

// PutFileCell returns a FileCell to the pool.
func PutFileCell(cell FileCell) {
	cell[FieldRuleName] = nil
	cell["Name"] = nil
	cell["Bytes"] = nil
	fileCellPool.Put(cell)
}
