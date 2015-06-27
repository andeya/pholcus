// 数据存储单元
package collector

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
