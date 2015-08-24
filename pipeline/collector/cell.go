package collector

import (
	"io"
)

// 数据存储单元
type DataCell map[string]interface{}

func NewDataCell(ruleName string, data map[string]interface{}, url string, parentUrl string, downloadTime string) DataCell {
	return DataCell{
		"RuleName":     ruleName,  //规定Data中的key
		"Data":         data,      //数据存储,key须与Rule的Fields保持一致
		"Url":          url,       //用于索引
		"ParentUrl":    parentUrl, //DataCell的上级url
		"DownloadTime": downloadTime,
	}
}

// 文件存储单元
type FileCell map[string]interface{}

// FileCell存储的完整文件名为： file/"Dir"/"RuleName"/"time"/"Name"

func NewFileCell(ruleName, name string, body io.ReadCloser) FileCell {
	return FileCell{
		"RuleName": ruleName, //存储路径中的一部分
		"Name":     name,     //规定文件名
		"Body":     body,     //文件内容
	}
}
