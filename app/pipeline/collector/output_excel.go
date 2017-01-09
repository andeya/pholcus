package collector

import (
	"fmt"
	"os"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/common/xlsx"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

/************************ excel 输出 ***************************/
func init() {
	DataOutput["excel"] = func(self *Collector) (err error) {
		defer func() {
			if p := recover(); p != nil {
				err = fmt.Errorf("%v", p)
			}
		}()

		var (
			file   *xlsx.File
			row    *xlsx.Row
			cell   *xlsx.Cell
			sheets = make(map[string]*xlsx.Sheet)
		)

		// 创建文件
		file = xlsx.NewFile()

		// 添加分类数据工作表
		for _, datacell := range self.dataDocker {
			var subNamespace = util.FileNameReplace(self.subNamespace(datacell))
			if _, ok := sheets[subNamespace]; !ok {
				// 添加工作表
				sheet, err := file.AddSheet(subNamespace)
				if err != nil {
					logs.Log.Error("%v", err)
					continue
				}
				sheets[subNamespace] = sheet
				// 写入表头
				row = sheets[subNamespace].AddRow()
				for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
					row.AddCell().Value = title
				}
				if self.Spider.OutDefaultField() {
					row.AddCell().Value = "当前链接"
					row.AddCell().Value = "上级链接"
					row.AddCell().Value = "下载时间"
				}
			}

			row = sheets[subNamespace].AddRow()
			for _, title := range self.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				cell = row.AddCell()
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					cell.Value = v
				} else {
					cell.Value = util.JsonString(vd[title])
				}
			}
			if self.Spider.OutDefaultField() {
				row.AddCell().Value = datacell["Url"].(string)
				row.AddCell().Value = datacell["ParentUrl"].(string)
				row.AddCell().Value = datacell["DownloadTime"].(string)
			}
		}
		folder := config.TEXT_DIR + "/" + cache.StartTime.Format("2006-01-02 150405")
		filename := fmt.Sprintf("%v/%v__%v-%v.xlsx", folder, util.FileNameReplace(self.namespace()), self.sum[0], self.sum[1])

		// 创建/打开目录
		f2, err := os.Stat(folder)
		if err != nil || !f2.IsDir() {
			if err := os.MkdirAll(folder, 0777); err != nil {
				logs.Log.Error("Error: %v\n", err)
			}
		}

		// 保存文件
		err = file.Save(filename)
		return
	}
}
