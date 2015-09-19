package collector

import (
	"fmt"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/tealeg/xlsx"
	"os"
)

/************************ excel 输出 ***************************/
func init() {
	Output["excel"] = func(self *Collector, dataIndex int) {
		defer func() {
			if err := recover(); err != nil {
				logs.Log.Error("%v", err)
			}
		}()

		var file *xlsx.File
		var sheet *xlsx.Sheet
		var row *xlsx.Row
		var cell *xlsx.Cell
		var err error

		folder1 := "result/data"
		folder2 := folder1 + "/" + self.startTime.Format("2006年01月02日 15时04分05秒")
		filename := folder2 + "/" + util.FileNameReplace(self.Spider.GetName()+"_"+self.Spider.GetKeyword()+" "+fmt.Sprintf("%v", self.sum[0])+"-"+fmt.Sprintf("%v", self.sum[1])) + ".xlsx"

		// 创建文件
		file = xlsx.NewFile()

		// 添加分类数据工作表
		for Name, Rule := range self.GetRules() {
			// 跳过不输出的数据
			if len(Rule.GetOutFeild()) == 0 {
				continue
			}
			// 添加工作表
			sheet = file.AddSheet(util.ExcelSheetNameReplace(Name))
			// 写入表头
			row = sheet.AddRow()
			for _, title := range Rule.GetOutFeild() {
				cell = row.AddCell()
				cell.Value = title
			}
			cell = row.AddCell()
			cell.Value = "当前链接"
			cell = row.AddCell()
			cell.Value = "上级链接"
			cell = row.AddCell()
			cell.Value = "下载时间"

			num := 0 //小计
			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				if datacell["RuleName"].(string) == Name {
					row = sheet.AddRow()
					for _, title := range Rule.GetOutFeild() {
						cell = row.AddCell()
						vd := datacell["Data"].(map[string]interface{})
						if v, ok := vd[title].(string); ok || vd[title] == nil {
							cell.Value = v
						} else {
							cell.Value = util.JsonString(vd[title])
						}
					}
					cell = row.AddCell()
					cell.Value = datacell["Url"].(string)
					cell = row.AddCell()
					cell.Value = datacell["ParentUrl"].(string)
					cell = row.AddCell()
					cell.Value = datacell["DownloadTime"].(string)
					num++
				}
			}
		}

		// 创建/打开目录
		f2, err := os.Stat(folder2)
		if err != nil || !f2.IsDir() {
			if err := os.MkdirAll(folder2, 0777); err != nil {
				logs.Log.Error("Error: %v\n", err)
			}
		}

		// 保存文件
		err = file.Save(filename)

		if err != nil {
			logs.Log.Error("%v", err)
		}

	}
}
