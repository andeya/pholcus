package collector

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/logs"
)

/************************ CSV 输出 ***************************/
func init() {
	Output["csv"] = func(self *Collector, dataIndex int) {
		defer func() {
			if err := recover(); err != nil {
				logs.Log.Error("%v", err)
			}
		}()

		folder1 := "result/data"
		folder2 := folder1 + "/" + self.startTime.Format("2006年01月02日 15时04分05秒")
		filenameBase := folder2 + "/" + util.FileNameReplace(self.Spider.GetName()+"_"+self.Spider.GetKeyword()+" "+fmt.Sprintf("%v", self.sum[0])+"-"+fmt.Sprintf("%v", self.sum[1]))

		// 创建/打开目录
		f2, err := os.Stat(folder2)
		if err != nil || !f2.IsDir() {
			if err := os.MkdirAll(folder2, 0777); err != nil {
				logs.Log.Error("Error: %v\n", err)
			}
		}

		// 按数据分类创建文件
		for Name, Rule := range self.GetRules() {
			// 跳过不输出的数据
			if len(Rule.GetOutFeild()) == 0 {
				continue
			}

			file, err := os.Create(filenameBase + " (" + util.FileNameReplace(Name) + ").csv")

			if err != nil {
				logs.Log.Error("%v", err)
				continue
			}

			file.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
			w := csv.NewWriter(file)
			th := Rule.GetOutFeild()
			th = append(th, []string{"当前链接", "上级链接", "下载时间"}...)
			w.Write(th)

			num := 0 //小计
			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				if datacell["RuleName"].(string) == Name {
					row := []string{}
					for _, title := range Rule.GetOutFeild() {
						vd := datacell["Data"].(map[string]interface{})
						if v, ok := vd[title].(string); ok || vd[title] == nil {
							row = append(row, v)
						} else {
							row = append(row, util.JsonString(vd[title]))
						}
					}

					row = append(row, datacell["Url"].(string))
					row = append(row, datacell["ParentUrl"].(string))
					row = append(row, datacell["DownloadTime"].(string))
					w.Write(row)

					num++
				}
			}
			// 发送缓存数据流
			w.Flush()
			// 关闭文件
			file.Close()
			// 输出报告
			// logs.Log.Critical("[任务：%v | 关键词：%v | 小类：%v] 输出 %v 条数据！！！\n", self.Spider.GetName(), self.Spider.GetKeyword(), Name, num)
		}
	}
}
